package browser

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/langowarny/lango/internal/logging"
)

var logger = logging.SubsystemSugar("tool.browser")

// ErrBrowserPanic is returned when a rod/CDP panic is recovered.
var ErrBrowserPanic = errors.New("browser panic recovered")

// safeRodCall wraps a rod call, converting any panic into an error.
func safeRodCall(fn func() error) (retErr error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("rod panic recovered", "panic", r)
			retErr = fmt.Errorf("%w: %v", ErrBrowserPanic, r)
		}
	}()
	return fn()
}

// safeRodCallValue wraps a rod call that returns a value, converting any panic into an error.
func safeRodCallValue[T any](fn func() (T, error)) (ret T, retErr error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorw("rod panic recovered", "panic", r)
			retErr = fmt.Errorf("%w: %v", ErrBrowserPanic, r)
		}
	}()
	return fn()
}

// Config holds browser tool configuration
type Config struct {
	Headless       bool
	BrowserBin     string
	SessionTimeout time.Duration
}

// Tool provides browser automation
type Tool struct {
	config   Config
	browser  *rod.Browser
	initMu   sync.Mutex
	initDone bool
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents a browser session with a page
type Session struct {
	ID        string
	Page      *rod.Page
	CreatedAt time.Time
}

// ScreenshotResult holds screenshot data
type ScreenshotResult struct {
	Data     string `json:"data"` // base64 encoded
	MimeType string `json:"mimeType"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
}

// ElementInfo holds element information
type ElementInfo struct {
	TagName   string `json:"tagName"`
	ID        string `json:"id,omitempty"`
	ClassName string `json:"className,omitempty"`
	InnerText string `json:"innerText,omitempty"`
	InnerHTML string `json:"innerHTML,omitempty"`
	Href      string `json:"href,omitempty"`
	Value     string `json:"value,omitempty"`
}

// New creates a new browser tool
func New(cfg Config) (*Tool, error) {
	if cfg.SessionTimeout == 0 {
		cfg.SessionTimeout = 5 * time.Minute
	}

	return &Tool{
		config:   cfg,
		sessions: make(map[string]*Session),
	}, nil
}

// ensureBrowser lazily initializes the browser with mutex-based guard.
// On failure, initDone remains false so the next call retries initialization.
func (t *Tool) ensureBrowser() error {
	t.initMu.Lock()
	defer t.initMu.Unlock()
	if t.initDone {
		return nil
	}
	if err := t.initBrowser(); err != nil {
		return err
	}
	t.initDone = true
	return nil
}

func (t *Tool) initBrowser() error {
	l := launcher.New().Headless(t.config.Headless)

	bin := t.config.BrowserBin
	if bin == "" {
		if found, has := launcher.LookPath(); has {
			bin = found
		}
	}
	if bin != "" {
		l = l.Bin(bin)
	}

	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("launch browser: %w", err)
	}

	b := rod.New().ControlURL(url)
	if err := b.Connect(); err != nil {
		return fmt.Errorf("connect browser: %w", err)
	}

	t.browser = b
	logger.Infow("browser launched", "headless", t.config.Headless, "bin", bin)
	return nil
}

// NewSession creates a new browser session
func (t *Tool) NewSession() (string, error) {
	if err := t.ensureBrowser(); err != nil {
		return "", err
	}

	page, err := safeRodCallValue(func() (*rod.Page, error) {
		return t.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	})
	if err != nil {
		return "", fmt.Errorf("create page: %w", err)
	}

	id := fmt.Sprintf("session-%d", time.Now().UnixNano())
	sess := &Session{
		ID:        id,
		Page:      page,
		CreatedAt: time.Now(),
	}

	t.mu.Lock()
	t.sessions[id] = sess
	t.mu.Unlock()

	logger.Infow("session created", "id", id)
	return id, nil
}

// CloseSession closes a browser session
func (t *Tool) CloseSession(sessionID string) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	session, ok := t.sessions[sessionID]
	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	session.Page.Close()
	delete(t.sessions, sessionID)

	logger.Infow("session closed", "id", sessionID)
	return nil
}

// HasSession reports whether a session with the given ID exists.
func (t *Tool) HasSession(sessionID string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, ok := t.sessions[sessionID]
	return ok
}

// getSession retrieves a session by ID
func (t *Tool) getSession(sessionID string) (*Session, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	session, ok := t.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	return session, nil
}

// Navigate navigates to a URL
func (t *Tool) Navigate(ctx context.Context, sessionID, url string) error {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	if err := safeRodCall(func() error {
		if err := sess.Page.Navigate(url); err != nil {
			return fmt.Errorf("navigation: %w", err)
		}
		if err := sess.Page.WaitLoad(); err != nil {
			return fmt.Errorf("wait load: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.Infow("navigated", "session", sessionID, "url", url)
	return nil
}

// Screenshot captures a screenshot
func (t *Tool) Screenshot(sessionID string, fullPage bool) (*ScreenshotResult, error) {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	data, err := safeRodCallValue(func() ([]byte, error) {
		return sess.Page.Screenshot(fullPage, nil)
	})
	if err != nil {
		return nil, fmt.Errorf("screenshot: %w", err)
	}

	return &ScreenshotResult{
		Data:     base64.StdEncoding.EncodeToString(data),
		MimeType: "image/png",
	}, nil
}

// Click clicks on an element
func (t *Tool) Click(ctx context.Context, sessionID, selector string) error {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	if err := safeRodCall(func() error {
		el, err := sess.Page.Element(selector)
		if err != nil {
			return fmt.Errorf("element not found: %s", selector)
		}
		if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
			return fmt.Errorf("click: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.Infow("clicked", "session", sessionID, "selector", selector)
	return nil
}

// Type types text into an element
func (t *Tool) Type(ctx context.Context, sessionID, selector, text string) error {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	if err := safeRodCall(func() error {
		el, err := sess.Page.Element(selector)
		if err != nil {
			return fmt.Errorf("element not found: %s", selector)
		}
		if err := el.Input(text); err != nil {
			return fmt.Errorf("type: %w", err)
		}
		return nil
	}); err != nil {
		return err
	}

	logger.Infow("typed", "session", sessionID, "selector", selector, "length", len(text))
	return nil
}

// GetText gets text content of an element
func (t *Tool) GetText(sessionID, selector string) (string, error) {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return "", err
	}

	return safeRodCallValue(func() (string, error) {
		el, err := sess.Page.Element(selector)
		if err != nil {
			return "", fmt.Errorf("element not found: %s", selector)
		}
		text, err := el.Text()
		if err != nil {
			return "", fmt.Errorf("get text: %w", err)
		}
		return text, nil
	})
}

// GetSnapshot returns basic page info (title and snippet)
func (t *Tool) GetSnapshot(sessionID string) (map[string]string, error) {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	return safeRodCallValue(func() (map[string]string, error) {
		info, err := sess.Page.Info()
		if err != nil {
			return nil, fmt.Errorf("get page info: %w", err)
		}

		body, err := sess.Page.Element("body")
		text := ""
		if err == nil {
			text, _ = body.Text()
		}

		if len(text) > 1000 {
			text = text[:1000] + "..."
		}

		return map[string]string{
			"title":   info.Title,
			"url":     info.URL,
			"snippet": text,
		}, nil
	})
}

// GetElementInfo gets information about an element
func (t *Tool) GetElementInfo(sessionID, selector string) (*ElementInfo, error) {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	return safeRodCallValue(func() (*ElementInfo, error) {
		el, err := sess.Page.Element(selector)
		if err != nil {
			return nil, fmt.Errorf("element not found: %s", selector)
		}

		tagName, _ := el.Eval(`() => this.tagName`)
		id, _ := el.Eval(`() => this.id`)
		className, _ := el.Eval(`() => this.className`)
		innerText, _ := el.Text()
		href, _ := el.Eval(`() => this.href || ""`)
		value, _ := el.Eval(`() => this.value || ""`)

		return &ElementInfo{
			TagName:   tagName.Value.String(),
			ID:        id.Value.String(),
			ClassName: className.Value.String(),
			InnerText: innerText,
			Href:      href.Value.String(),
			Value:     value.Value.String(),
		}, nil
	})
}

// Eval executes JavaScript on the page
func (t *Tool) Eval(sessionID, script string) (interface{}, error) {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	result, err := safeRodCallValue(func() (interface{}, error) {
		res, err := sess.Page.Eval(script)
		if err != nil {
			return nil, fmt.Errorf("eval: %w", err)
		}
		return res.Value.Val(), nil
	})
	if err != nil {
		return nil, err
	}

	logger.Infow("eval executed", "session", sessionID)
	return result, nil
}

// WaitForSelector waits for an element to appear
func (t *Tool) WaitForSelector(ctx context.Context, sessionID, selector string, timeout time.Duration) error {
	sess, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	return safeRodCall(func() error {
		sess.Page.Timeout(timeout)
		_, err := sess.Page.Element(selector)
		if err != nil {
			return fmt.Errorf("timeout waiting for: %s", selector)
		}
		return nil
	})
}

// Close closes all sessions and the browser
func (t *Tool) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id, sess := range t.sessions {
		_ = safeRodCall(func() error { sess.Page.Close(); return nil })
		delete(t.sessions, id)
	}

	if t.browser != nil {
		_ = safeRodCall(func() error { t.browser.Close(); return nil })
		t.browser = nil
	}

	// Reset so browser can be re-initialized if needed
	t.initMu.Lock()
	t.initDone = false
	t.initMu.Unlock()

	logger.Info("browser closed")
	return nil
}
