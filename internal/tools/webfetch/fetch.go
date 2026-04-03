package webfetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout    = 30 * time.Second
	defaultMaxLength  = 5000
	defaultMaxBody    = 5 * 1024 * 1024 // 5 MB max download
	defaultUserAgent  = "Lango/1.0 (Web Fetch Tool)"
	maxRedirects      = 5

	// ModeText extracts readable text from the HTML.
	ModeText = "text"
	// ModeHTML returns raw HTML truncated to max length.
	ModeHTML = "html"
	// ModeMarkdown returns simplified markdown.
	ModeMarkdown = "markdown"
)

// ErrBlockedURL is returned when a URL targets an internal or private network address.
var ErrBlockedURL = errors.New("URL targets a blocked internal/private network address")

// privateNetworks defines CIDR ranges considered internal/private.
var privateNetworks = []net.IPNet{
	{IP: net.IP{10, 0, 0, 0}, Mask: net.CIDRMask(8, 32)},
	{IP: net.IP{172, 16, 0, 0}, Mask: net.CIDRMask(12, 32)},
	{IP: net.IP{192, 168, 0, 0}, Mask: net.CIDRMask(16, 32)},
	{IP: net.IP{169, 254, 0, 0}, Mask: net.CIDRMask(16, 32)},
	{IP: net.IP{127, 0, 0, 0}, Mask: net.CIDRMask(8, 32)},
}

// FetchResult holds the extracted content from a fetched web page.
type FetchResult struct {
	URL           string `json:"url"`
	Title         string `json:"title"`
	Content       string `json:"content"`
	ContentLength int    `json:"content_length"`
	Truncated     bool   `json:"truncated"`
}

// Fetch downloads a web page and extracts content in the specified mode.
// Supported modes: "text" (default), "html" (raw HTML), "markdown" (simplified).
// maxLength controls the maximum character length of the returned content.
// When p2pSafe is true, each redirect target is validated before following.
func Fetch(ctx context.Context, rawURL string, mode string, maxLength int, p2pSafe bool) (*FetchResult, error) {
	if rawURL == "" {
		return nil, errors.New("empty URL")
	}
	if mode == "" {
		mode = ModeText
	}
	if maxLength <= 0 {
		maxLength = defaultMaxLength
	}

	// Validate mode.
	switch mode {
	case ModeText, ModeHTML, ModeMarkdown:
	default:
		return nil, fmt.Errorf("unsupported mode %q: use text, html, or markdown", mode)
	}

	// Ensure URL has a scheme.
	if !strings.Contains(rawURL, "://") {
		rawURL = "https://" + rawURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL: %w", err)
	}
	if parsed.Host == "" {
		return nil, errors.New("URL missing host")
	}

	client := &http.Client{
		Timeout: defaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= maxRedirects {
				return fmt.Errorf("stopped after %d redirects", maxRedirects)
			}
			// Validate each redirect target before following in P2P context.
			if p2pSafe {
				if err := ValidateURLForP2P(req.URL.String()); err != nil {
					return fmt.Errorf("redirect blocked: %w", err)
				}
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,*/*;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch %q: %w", rawURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d for %q", resp.StatusCode, rawURL)
	}

	// Limit body read to prevent memory exhaustion.
	bodyReader := io.LimitReader(resp.Body, int64(defaultMaxBody))

	finalURL := resp.Request.URL.String()

	switch mode {
	case ModeHTML:
		return extractHTMLMode(bodyReader, finalURL, maxLength)
	case ModeMarkdown:
		return extractMarkdownMode(bodyReader, finalURL, maxLength)
	default:
		return extractTextMode(bodyReader, finalURL, maxLength)
	}
}

// extractTextMode reads HTML and returns clean text.
func extractTextMode(r io.Reader, finalURL string, maxLength int) (*FetchResult, error) {
	title, body, err := extractText(r)
	if err != nil {
		return nil, err
	}
	return buildResult(finalURL, title, body, maxLength), nil
}

// extractMarkdownMode reads HTML and returns simplified markdown.
func extractMarkdownMode(r io.Reader, finalURL string, maxLength int) (*FetchResult, error) {
	title, body, err := extractMarkdown(r)
	if err != nil {
		return nil, err
	}
	return buildResult(finalURL, title, body, maxLength), nil
}

// extractHTMLMode reads raw HTML and returns it truncated.
func extractHTMLMode(r io.Reader, finalURL string, maxLength int) (*FetchResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}
	body := string(data)
	// Try to find a title from the HTML.
	title, _, _ := extractText(strings.NewReader(body))
	return buildResult(finalURL, title, body, maxLength), nil
}

// buildResult constructs a FetchResult, truncating content if needed.
func buildResult(url, title, content string, maxLength int) *FetchResult {
	truncated := false
	if len(content) > maxLength {
		content = content[:maxLength]
		truncated = true
	}
	return &FetchResult{
		URL:           url,
		Title:         title,
		Content:       content,
		ContentLength: len(content),
		Truncated:     truncated,
	}
}

// ValidateURLForP2P checks that a URL is safe in a P2P context.
// It blocks file:// schemes and URLs resolving to internal/private network addresses.
func ValidateURLForP2P(rawURL string) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse URL: %w", err)
	}

	if strings.EqualFold(parsed.Scheme, "file") {
		return fmt.Errorf("%w: file:// scheme is not allowed", ErrBlockedURL)
	}

	hostname := parsed.Hostname()
	if hostname == "" {
		return fmt.Errorf("%w: empty hostname", ErrBlockedURL)
	}

	lower := strings.ToLower(hostname)
	if lower == "localhost" {
		return fmt.Errorf("%w: localhost is not allowed", ErrBlockedURL)
	}
	if lower == "::1" {
		return fmt.Errorf("%w: IPv6 loopback is not allowed", ErrBlockedURL)
	}

	ip := net.ParseIP(hostname)
	if ip != nil {
		return checkIPPrivate(ip, hostname)
	}

	// Hostname is not an IP literal — resolve via DNS and check all results.
	ips, err := net.LookupIP(hostname)
	if err == nil {
		for _, resolved := range ips {
			if err := checkIPPrivate(resolved, hostname); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkIPPrivate returns an error if ip falls within a private/loopback range.
func checkIPPrivate(ip net.IP, label string) error {
	if ip.IsLoopback() {
		return fmt.Errorf("%w: loopback address is not allowed", ErrBlockedURL)
	}
	for _, cidr := range privateNetworks {
		if cidr.Contains(ip) {
			return fmt.Errorf("%w: %s resolves to a private network address", ErrBlockedURL, label)
		}
	}
	return nil
}
