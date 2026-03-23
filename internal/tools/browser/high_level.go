package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

const (
	defaultSearchEndpoint     = "https://duckduckgo.com/html/"
	defaultLinkLimit          = 8
	defaultActionLimit        = 8
	defaultObservationLimit   = 10
	defaultSearchResultsLimit = 5
	defaultHeadingLimit       = 8
	maxExtractionLimit        = 20
)

// PageLink is a structured link extracted from the current page.
type PageLink struct {
	Text     string `json:"text"`
	URL      string `json:"url"`
	Selector string `json:"selector,omitempty"`
}

// ActionCandidate is an actionable element discovered on the current page.
type ActionCandidate struct {
	Selector    string `json:"selector"`
	Tag         string `json:"tag"`
	Type        string `json:"type,omitempty"`
	Role        string `json:"role,omitempty"`
	Text        string `json:"text,omitempty"`
	Href        string `json:"href,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
}

// PageSnapshot is a structured summary of the current page.
type PageSnapshot struct {
	Title    string            `json:"title"`
	URL      string            `json:"url"`
	Snippet  string            `json:"snippet"`
	Headings []string          `json:"headings,omitempty"`
	Links    []PageLink        `json:"links,omitempty"`
	Actions  []ActionCandidate `json:"actions,omitempty"`
}

// LinksResult contains extracted links from the current page.
type LinksResult struct {
	Title string     `json:"title"`
	URL   string     `json:"url"`
	Links []PageLink `json:"links,omitempty"`
}

// ArticleResult contains structured article-like content from the current page.
type ArticleResult struct {
	Title    string   `json:"title"`
	URL      string   `json:"url"`
	Headings []string `json:"headings,omitempty"`
	Content  string   `json:"content"`
}

// SearchResult is a structured web search result item.
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet,omitempty"`
}

// SearchResponse is a structured response for browser-based web search.
type SearchResponse struct {
	Query   string         `json:"query,omitempty"`
	Title   string         `json:"title"`
	URL     string         `json:"url"`
	Results []SearchResult `json:"results,omitempty"`
}

// Snapshot returns a structured summary of the current page.
func (t *Tool) Snapshot(sessionID string, linkLimit, actionLimit int) (*PageSnapshot, error) {
	linkLimit = clampLimit(linkLimit, defaultLinkLimit)
	actionLimit = clampLimit(actionLimit, defaultActionLimit)

	var out PageSnapshot
	if err := t.evalInto(sessionID, snapshotScript(linkLimit, actionLimit), &out); err != nil {
		return nil, fmt.Errorf("snapshot: %w", err)
	}
	return &out, nil
}

// Observe returns actionable elements from the current page.
func (t *Tool) Observe(sessionID string, limit int) ([]ActionCandidate, error) {
	snapshot, err := t.Snapshot(sessionID, defaultLinkLimit, clampLimit(limit, defaultObservationLimit))
	if err != nil {
		return nil, err
	}
	return snapshot.Actions, nil
}

// Extract returns structured data from the current page.
func (t *Tool) Extract(sessionID, mode string, limit int) (interface{}, error) {
	switch mode {
	case "", "summary":
		return t.Snapshot(sessionID, clampLimit(limit, defaultLinkLimit), clampLimit(limit, defaultActionLimit))
	case "links":
		snapshot, err := t.Snapshot(sessionID, clampLimit(limit, defaultLinkLimit), defaultActionLimit)
		if err != nil {
			return nil, err
		}
		return &LinksResult{
			Title: snapshot.Title,
			URL:   snapshot.URL,
			Links: snapshot.Links,
		}, nil
	case "article":
		return t.extractArticle(sessionID)
	case "search_results":
		return t.extractSearchResults(sessionID, limit)
	default:
		return nil, fmt.Errorf("unknown extract mode: %s", mode)
	}
}

// Search performs browser-native web search and returns structured results.
func (t *Tool) Search(ctx context.Context, sessionID, query string, limit int) (*SearchResponse, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	searchURL := buildSearchURL(query)
	if err := t.Navigate(ctx, sessionID, searchURL); err != nil {
		return nil, err
	}

	out, err := t.extractSearchResults(sessionID, limit)
	if err != nil {
		return nil, err
	}
	out.Query = query
	return out, nil
}

func (t *Tool) extractArticle(sessionID string) (*ArticleResult, error) {
	var out ArticleResult
	if err := t.evalInto(sessionID, articleScript(), &out); err != nil {
		return nil, fmt.Errorf("extract article: %w", err)
	}
	return &out, nil
}

func (t *Tool) extractSearchResults(sessionID string, limit int) (*SearchResponse, error) {
	var out SearchResponse
	if err := t.evalInto(sessionID, searchResultsScript(clampLimit(limit, defaultSearchResultsLimit)), &out); err != nil {
		return nil, fmt.Errorf("extract search results: %w", err)
	}
	return &out, nil
}

func (t *Tool) evalInto(sessionID, script string, out interface{}) error {
	raw, err := t.Eval(sessionID, script)
	if err != nil {
		return err
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return fmt.Errorf("marshal eval result: %w", err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode eval result: %w", err)
	}
	return nil
}

func clampLimit(limit, def int) int {
	if limit <= 0 {
		limit = def
	}
	if limit > maxExtractionLimit {
		return maxExtractionLimit
	}
	return limit
}

func buildSearchURL(query string) string {
	u, err := url.Parse(defaultSearchEndpoint)
	if err != nil {
		return defaultSearchEndpoint
	}

	values := u.Query()
	values.Set("q", query)
	u.RawQuery = values.Encode()
	return u.String()
}

func snapshotScript(linkLimit, actionLimit int) string {
	return fmt.Sprintf(`() => {
		const normalize = (value) => String(value || "").replace(/\s+/g, " ").trim();
		const absURL = (value) => {
			try {
				return new URL(value, window.location.href).href;
			} catch (error) {
				return "";
			}
		};
		const isVisible = (el) => {
			if (!(el instanceof Element)) {
				return false;
			}
			const style = window.getComputedStyle(el);
			if (!style || style.visibility === "hidden" || style.display === "none") {
				return false;
			}
			const rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		};
		const uniqueBy = (items, keyFn) => {
			const seen = new Set();
			return items.filter((item) => {
				const key = keyFn(item);
				if (!key || seen.has(key)) {
					return false;
				}
				seen.add(key);
				return true;
			});
		};
		const selectorFor = (el) => {
			if (!(el instanceof Element)) {
				return "";
			}
			if (el.id) {
				return "#" + CSS.escape(el.id);
			}
			const parts = [];
			let node = el;
			while (node && node.nodeType === Node.ELEMENT_NODE && parts.length < 5) {
				let part = node.tagName.toLowerCase();
				if (node.parentElement) {
					const siblings = Array.from(node.parentElement.children)
						.filter((child) => child.tagName === node.tagName);
					if (siblings.length > 1) {
						part += ":nth-of-type(" + (siblings.indexOf(node) + 1) + ")";
					}
				}
				parts.unshift(part);
				if (node.parentElement && node.parentElement.id) {
					parts.unshift("#" + CSS.escape(node.parentElement.id));
					break;
				}
				node = node.parentElement;
			}
			return parts.join(" > ");
		};

		const bodyText = normalize(document.body ? document.body.innerText : "");
		const headings = uniqueBy(
			Array.from(document.querySelectorAll("h1,h2,h3"))
				.map((el) => normalize(el.innerText))
				.filter(Boolean),
			(item) => item
		).slice(0, %d);

		const links = uniqueBy(
			Array.from(document.querySelectorAll("a[href]"))
				.filter((el) => isVisible(el))
				.map((el) => {
					const href = absURL(el.getAttribute("href") || el.href);
					return {
						text: normalize(el.innerText || el.getAttribute("aria-label") || el.title),
						url: href,
						selector: selectorFor(el)
					};
				})
				.filter((item) => item.text && /^https?:\/\//i.test(item.url)),
			(item) => item.url
		).slice(0, %d);

		const actions = uniqueBy(
			Array.from(document.querySelectorAll("a[href],button,input,textarea,select,[role='button'],[role='link']"))
				.filter((el) => isVisible(el))
				.map((el) => ({
					selector: selectorFor(el),
					tag: (el.tagName || "").toLowerCase(),
					type: normalize(el.getAttribute("type")),
					role: normalize(el.getAttribute("role")),
					text: normalize(el.innerText || el.getAttribute("aria-label") || el.title || el.value),
					href: absURL(el.getAttribute("href") || ""),
					placeholder: normalize(el.getAttribute("placeholder"))
				}))
				.filter((item) => item.selector && (item.text || item.href || item.placeholder)),
			(item) => item.selector
		).slice(0, %d);

		return {
			title: document.title || "",
			url: window.location.href,
			snippet: bodyText.slice(0, 1000),
			headings: headings,
			links: links,
			actions: actions
		};
	}`, defaultHeadingLimit, linkLimit, actionLimit)
}

func articleScript() string {
	return fmt.Sprintf(`() => {
		const normalize = (value) => String(value || "").replace(/\s+/g, " ").trim();
		const uniqueBy = (items, keyFn) => {
			const seen = new Set();
			return items.filter((item) => {
				const key = keyFn(item);
				if (!key || seen.has(key)) {
					return false;
				}
				seen.add(key);
				return true;
			});
		};

		const root = document.querySelector("article") || document.querySelector("main") || document.body;
		const headings = uniqueBy(
			Array.from(root.querySelectorAll("h1,h2,h3"))
				.map((el) => normalize(el.innerText))
				.filter(Boolean),
			(item) => item
		).slice(0, %d);

		return {
			title: document.title || "",
			url: window.location.href,
			headings: headings,
			content: normalize(root ? root.innerText : "").slice(0, 5000)
		};
	}`, defaultHeadingLimit)
}

func searchResultsScript(limit int) string {
	return fmt.Sprintf(`() => {
		const normalize = (value) => String(value || "").replace(/\s+/g, " ").trim();
		const absURL = (value) => {
			try {
				return new URL(value, window.location.href).href;
			} catch (error) {
				return "";
			}
		};
		const isVisible = (el) => {
			if (!(el instanceof Element)) {
				return false;
			}
			const style = window.getComputedStyle(el);
			if (!style || style.visibility === "hidden" || style.display === "none") {
				return false;
			}
			const rect = el.getBoundingClientRect();
			return rect.width > 0 && rect.height > 0;
		};
		const uniqueBy = (items, keyFn) => {
			const seen = new Set();
			return items.filter((item) => {
				const key = keyFn(item);
				if (!key || seen.has(key)) {
					return false;
				}
				seen.add(key);
				return true;
			});
		};
		const pickText = (root, selectors) => {
			for (const selector of selectors) {
				const el = root.querySelector(selector);
				if (!el) {
					continue;
				}
				const text = normalize(el.innerText || el.textContent);
				if (text) {
					return text;
				}
			}
			return "";
		};

		let containers = [];
		const selectors = [
			"article[data-testid='result']",
			"[data-testid='result']",
			".result",
			".result__body",
			"li.b_algo",
			".g"
		];
		for (const selector of selectors) {
			containers = containers.concat(Array.from(document.querySelectorAll(selector)).filter(isVisible));
		}

		const uniqueContainers = [];
		const seenContainers = new Set();
		for (const el of containers) {
			if (seenContainers.has(el)) {
				continue;
			}
			seenContainers.add(el);
			uniqueContainers.push(el);
		}

		let results = uniqueContainers.map((root) => {
			const anchor = root.querySelector("a[href]");
			if (!anchor) {
				return null;
			}
			const resultURL = absURL(anchor.getAttribute("href") || anchor.href);
			const title = normalize(
				anchor.innerText ||
				pickText(root, ["h1", "h2", "h3"]) ||
				anchor.getAttribute("aria-label") ||
				anchor.getAttribute("title")
			);
			const snippet = pickText(root, [
				".result__snippet",
				".snippet",
				".b_caption p",
				"p"
			]);
			if (!title || !/^https?:\/\//i.test(resultURL)) {
				return null;
			}
			return {
				title: title,
				url: resultURL,
				snippet: snippet
			};
		}).filter(Boolean);

		results = uniqueBy(results, (item) => item.url);

		if (results.length === 0) {
			results = uniqueBy(
				Array.from(document.querySelectorAll("a[href]"))
					.filter((el) => isVisible(el))
					.map((el) => ({
						title: normalize(el.innerText || el.getAttribute("aria-label") || el.title),
						url: absURL(el.getAttribute("href") || el.href),
						snippet: ""
					}))
					.filter((item) => item.title.length >= 12 && /^https?:\/\//i.test(item.url)),
				(item) => item.url
			);
		}

		return {
			title: document.title || "",
			url: window.location.href,
			results: results.slice(0, %d)
		};
	}`, limit)
}
