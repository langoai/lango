package websearch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"

	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/langoai/lango/internal/tools/browser"
)

const (
	defaultLimit     = 5
	maxLimit         = 20
	httpTimeout      = 15 * time.Second
	defaultUserAgent = "Mozilla/5.0 (compatible; Lango/1.0)"
)

// searchEndpoint is the DuckDuckGo HTML search URL prefix.
// It is a var so tests can swap it to a local httptest server.
var searchEndpoint = "https://html.duckduckgo.com/html/?q="

// SearchResult is a single web search result.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// Search queries DuckDuckGo's HTML endpoint and returns structured results.
// When the context carries a P2P request flag, result URLs are validated
// against private/internal network ranges before inclusion.
func Search(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	limit = clampLimit(limit)

	reqURL := searchEndpoint + url.QueryEscape(query)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", defaultUserAgent)

	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	results, err := parseResults(resp.Body, limit)
	if err != nil {
		return nil, fmt.Errorf("parse results: %w", err)
	}

	// In P2P context, filter out results pointing to private/internal URLs.
	if ctxkeys.IsP2PRequest(ctx) {
		results = filterP2PSafe(results)
	}

	return results, nil
}

// parseResults reads the DuckDuckGo HTML response and extracts search results.
// DuckDuckGo's HTML endpoint renders results as:
//
//	<div class="result results_links results_links_deep web-result">
//	  <h2 class="result__title">
//	    <a class="result__a" href="...">Title</a>
//	  </h2>
//	  <a class="result__snippet">Snippet text</a>
//	</div>
func parseResults(body interface{ Read([]byte) (int, error) }, limit int) ([]SearchResult, error) {
	doc, err := html.Parse(body)
	if err != nil {
		return nil, err
	}

	var results []SearchResult
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if len(results) >= limit {
			return
		}

		// Look for result container divs.
		if isResultDiv(n) {
			if sr, ok := extractResult(n); ok {
				results = append(results, sr)
			}
			return // Don't recurse into result containers.
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	return results, nil
}

// isResultDiv checks if a node is a DuckDuckGo result container.
func isResultDiv(n *html.Node) bool {
	if n.Type != html.ElementNode || n.Data != "div" {
		return false
	}
	cls := attrVal(n, "class")
	return strings.Contains(cls, "result") && strings.Contains(cls, "web-result")
}

// extractResult extracts title, URL, and snippet from a result container.
func extractResult(container *html.Node) (SearchResult, bool) {
	var sr SearchResult

	// Find the title link: <a class="result__a" ...>
	if a := findNode(container, func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "a" && strings.Contains(attrVal(n, "class"), "result__a")
	}); a != nil {
		sr.Title = normalizeText(textContent(a))
		sr.URL = attrVal(a, "href")
	}

	// Find snippet: <a class="result__snippet"> or <span class="result__snippet">
	if snippet := findNode(container, func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		return strings.Contains(attrVal(n, "class"), "result__snippet")
	}); snippet != nil {
		sr.Snippet = normalizeText(textContent(snippet))
	}

	if sr.Title == "" || sr.URL == "" {
		return sr, false
	}

	// Resolve relative URLs or DuckDuckGo redirect URLs.
	sr.URL = resolveURL(sr.URL)

	return sr, sr.URL != ""
}

// resolveURL cleans up DuckDuckGo's redirect URLs.
// DDG wraps result URLs like //duckduckgo.com/l/?uddg=<encoded_url>&...
func resolveURL(raw string) string {
	// Handle protocol-relative URLs.
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return ""
	}

	// Check if this is a DDG redirect URL.
	if strings.Contains(parsed.Host, "duckduckgo.com") && parsed.Path == "/l/" {
		if uddg := parsed.Query().Get("uddg"); uddg != "" {
			return uddg
		}
	}

	// Only return http(s) URLs.
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return ""
	}

	return raw
}

// filterP2PSafe removes results whose URLs target private/internal networks.
func filterP2PSafe(results []SearchResult) []SearchResult {
	safe := make([]SearchResult, 0, len(results))
	for _, r := range results {
		if err := browser.ValidateURLForP2P(r.URL); err == nil {
			safe = append(safe, r)
		}
	}
	return safe
}

// findNode recursively searches for the first node matching the predicate.
func findNode(n *html.Node, match func(*html.Node) bool) *html.Node {
	if match(n) {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findNode(c, match); found != nil {
			return found
		}
	}
	return nil
}

// textContent returns the concatenated text content of a node and its children.
func textContent(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		sb.WriteString(textContent(c))
	}
	return sb.String()
}

// attrVal returns the value of the named attribute, or "" if not found.
func attrVal(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if a.Key == name {
			return a.Val
		}
	}
	return ""
}

// normalizeText collapses whitespace and trims.
func normalizeText(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

// clampLimit bounds the result limit to [1, maxLimit] with a default.
func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
