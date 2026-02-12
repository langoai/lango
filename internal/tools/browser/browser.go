package browser

import (
	"context"
	"encoding/base64"
	"fmt"
	"sync"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/langowarny/lango/internal/logging"
)

var logger = logging.SubsystemSugar("tool.browser")

// Config holds browser tool configuration
type Config struct {
	Headless       bool
	SessionTimeout time.Duration
}

// Tool provides browser automation
type Tool struct {
	config   Config
	browser  *rod.Browser
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

// ensureBrowser lazily initializes the browser
func (t *Tool) ensureBrowser() error {
	if t.browser != nil {
		return nil
	}

	l := launcher.New().Headless(t.config.Headless)
	url, err := l.Launch()
	if err != nil {
		return fmt.Errorf("failed to launch browser: %w", err)
	}

	t.browser = rod.New().ControlURL(url)
	if err := t.browser.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	logger.Infow("browser launched", "headless", t.config.Headless)
	return nil
}

// NewSession creates a new browser session
func (t *Tool) NewSession() (string, error) {
	if err := t.ensureBrowser(); err != nil {
		return "", err
	}

	page, err := t.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}

	id := fmt.Sprintf("session-%d", time.Now().UnixNano())
	session := &Session{
		ID:        id,
		Page:      page,
		CreatedAt: time.Now(),
	}

	t.mu.Lock()
	t.sessions[id] = session
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
	session, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	if err := session.Page.Navigate(url); err != nil {
		return fmt.Errorf("navigation failed: %w", err)
	}

	// Wait for load
	if err := session.Page.WaitLoad(); err != nil {
		return fmt.Errorf("wait load failed: %w", err)
	}

	logger.Infow("navigated", "session", sessionID, "url", url)
	return nil
}

// Screenshot captures a screenshot
func (t *Tool) Screenshot(sessionID string, fullPage bool) (*ScreenshotResult, error) {
	session, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	var data []byte
	if fullPage {
		data, err = session.Page.Screenshot(true, nil)
	} else {
		data, err = session.Page.Screenshot(false, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("screenshot failed: %w", err)
	}

	return &ScreenshotResult{
		Data:     base64.StdEncoding.EncodeToString(data),
		MimeType: "image/png",
	}, nil
}

// Click clicks on an element
func (t *Tool) Click(ctx context.Context, sessionID, selector string) error {
	session, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	el, err := session.Page.Element(selector)
	if err != nil {
		return fmt.Errorf("element not found: %s", selector)
	}

	if err := el.Click(proto.InputMouseButtonLeft, 1); err != nil {
		return fmt.Errorf("click failed: %w", err)
	}

	logger.Infow("clicked", "session", sessionID, "selector", selector)
	return nil
}

// Type types text into an element
func (t *Tool) Type(ctx context.Context, sessionID, selector, text string) error {
	session, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	el, err := session.Page.Element(selector)
	if err != nil {
		return fmt.Errorf("element not found: %s", selector)
	}

	if err := el.Input(text); err != nil {
		return fmt.Errorf("type failed: %w", err)
	}

	logger.Infow("typed", "session", sessionID, "selector", selector, "length", len(text))
	return nil
}

// GetText gets text content of an element
func (t *Tool) GetText(sessionID, selector string) (string, error) {
	session, err := t.getSession(sessionID)
	if err != nil {
		return "", err
	}

	el, err := session.Page.Element(selector)
	if err != nil {
		return "", fmt.Errorf("element not found: %s", selector)
	}

	text, err := el.Text()
	if err != nil {
		return "", fmt.Errorf("get text failed: %w", err)
	}

	return text, nil
}

// GetSnapshot returns basic page info (title and snippet)
func (t *Tool) GetSnapshot(sessionID string) (map[string]string, error) {
	session, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	info, err := session.Page.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}

	body, err := session.Page.Element("body")
	text := ""
	if err == nil {
		text, _ = body.Text()
	}

	// Limit snippet
	if len(text) > 1000 {
		text = text[:1000] + "..."
	}

	return map[string]string{
		"title":   info.Title,
		"url":     info.URL,
		"snippet": text,
	}, nil
}

// GetElementInfo gets information about an element
func (t *Tool) GetElementInfo(sessionID, selector string) (*ElementInfo, error) {
	session, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	el, err := session.Page.Element(selector)
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
}

// Eval executes JavaScript on the page
func (t *Tool) Eval(sessionID, script string) (interface{}, error) {
	session, err := t.getSession(sessionID)
	if err != nil {
		return nil, err
	}

	result, err := session.Page.Eval(script)
	if err != nil {
		return nil, fmt.Errorf("eval failed: %w", err)
	}

	logger.Infow("eval executed", "session", sessionID)
	return result.Value.Val(), nil
}

// WaitForSelector waits for an element to appear
func (t *Tool) WaitForSelector(ctx context.Context, sessionID, selector string, timeout time.Duration) error {
	session, err := t.getSession(sessionID)
	if err != nil {
		return err
	}

	if timeout == 0 {
		timeout = 10 * time.Second
	}

	session.Page.Timeout(timeout)
	_, err = session.Page.Element(selector)
	if err != nil {
		return fmt.Errorf("timeout waiting for: %s", selector)
	}

	return nil
}

// Close closes all sessions and the browser
func (t *Tool) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	for id, session := range t.sessions {
		session.Page.Close()
		delete(t.sessions, id)
	}

	if t.browser != nil {
		t.browser.Close()
		t.browser = nil
	}

	logger.Info("browser closed")
	return nil
}
