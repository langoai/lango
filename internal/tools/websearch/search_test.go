package websearch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/langoai/lango/internal/ctxkeys"
)

// sampleDDGHTML is a minimal DuckDuckGo HTML response for testing.
const sampleDDGHTML = `<!DOCTYPE html>
<html>
<head><title>DuckDuckGo</title></head>
<body>
<div id="links">
  <div class="result results_links results_links_deep web-result">
    <h2 class="result__title">
      <a class="result__a" href="https://example.com/page1">Example Page One</a>
    </h2>
    <a class="result__snippet">This is the first result snippet.</a>
  </div>
  <div class="result results_links results_links_deep web-result">
    <h2 class="result__title">
      <a class="result__a" href="https://example.org/page2">Example Page Two</a>
    </h2>
    <a class="result__snippet">This is the second result snippet.</a>
  </div>
  <div class="result results_links results_links_deep web-result">
    <h2 class="result__title">
      <a class="result__a" href="https://example.net/page3">Example Page Three</a>
    </h2>
    <a class="result__snippet">This is the third result snippet.</a>
  </div>
</div>
</body>
</html>`

// sampleDDGWithRedirect contains a DDG redirect URL.
const sampleDDGWithRedirect = `<!DOCTYPE html>
<html>
<body>
<div id="links">
  <div class="result results_links results_links_deep web-result">
    <h2 class="result__title">
      <a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Freal.example.com%2Fpath&amp;rut=abc">Real Page</a>
    </h2>
    <a class="result__snippet">Redirected result.</a>
  </div>
</div>
</body>
</html>`

// sampleDDGWithPrivateURL contains a result pointing to a private IP.
const sampleDDGWithPrivateURL = `<!DOCTYPE html>
<html>
<body>
<div id="links">
  <div class="result results_links results_links_deep web-result">
    <h2 class="result__title">
      <a class="result__a" href="https://example.com/safe">Safe Page</a>
    </h2>
    <a class="result__snippet">Public result.</a>
  </div>
  <div class="result results_links results_links_deep web-result">
    <h2 class="result__title">
      <a class="result__a" href="http://192.168.1.1/admin">Router Admin</a>
    </h2>
    <a class="result__snippet">Private network page.</a>
  </div>
</div>
</body>
</html>`

const sampleDDGEmpty = `<!DOCTYPE html>
<html>
<body>
<div id="links">
  <div class="no-results">No results found.</div>
</div>
</body>
</html>`

// overrideEndpoint swaps the package-level search endpoint for testing.
func overrideEndpoint(url string) { searchEndpoint = url }

// restoreEndpoint restores the original search endpoint after testing.
func restoreEndpoint(orig string) { searchEndpoint = orig }

func TestSearch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if q == "" {
			http.Error(w, "missing query", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGHTML))
	}))
	defer ts.Close()

	// Override search endpoint for testing.
	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	ctx := context.Background()
	results, err := Search(ctx, "test query", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("got %d results, want 3", len(results))
	}

	tests := []struct {
		wantTitle   string
		wantURL     string
		wantSnippet string
	}{
		{
			wantTitle:   "Example Page One",
			wantURL:     "https://example.com/page1",
			wantSnippet: "This is the first result snippet.",
		},
		{
			wantTitle:   "Example Page Two",
			wantURL:     "https://example.org/page2",
			wantSnippet: "This is the second result snippet.",
		},
		{
			wantTitle:   "Example Page Three",
			wantURL:     "https://example.net/page3",
			wantSnippet: "This is the third result snippet.",
		},
	}

	for i, tt := range tests {
		if results[i].Title != tt.wantTitle {
			t.Errorf("result[%d].Title = %q, want %q", i, results[i].Title, tt.wantTitle)
		}
		if results[i].URL != tt.wantURL {
			t.Errorf("result[%d].URL = %q, want %q", i, results[i].URL, tt.wantURL)
		}
		if results[i].Snippet != tt.wantSnippet {
			t.Errorf("result[%d].Snippet = %q, want %q", i, results[i].Snippet, tt.wantSnippet)
		}
	}
}

func TestSearch_Limit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGHTML))
	}))
	defer ts.Close()

	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	ctx := context.Background()
	results, err := Search(ctx, "test", 2)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestSearch_EmptyQuery(t *testing.T) {
	_, err := Search(context.Background(), "", 5)
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestSearch_EmptyResults(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGEmpty))
	}))
	defer ts.Close()

	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	results, err := Search(context.Background(), "nothing", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestSearch_DDGRedirectURL(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGWithRedirect))
	}))
	defer ts.Close()

	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	results, err := Search(context.Background(), "redirect test", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}

	if results[0].URL != "https://real.example.com/path" {
		t.Errorf("URL = %q, want %q", results[0].URL, "https://real.example.com/path")
	}
}

func TestSearch_P2PFiltersPrivateURLs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGWithPrivateURL))
	}))
	defer ts.Close()

	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	ctx := ctxkeys.WithP2PRequest(context.Background())
	results, err := Search(ctx, "test p2p", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	// The private URL (192.168.1.1) should be filtered out.
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1 (private URL filtered)", len(results))
	}

	if results[0].URL != "https://example.com/safe" {
		t.Errorf("URL = %q, want %q", results[0].URL, "https://example.com/safe")
	}
}

func TestSearch_NonP2PKeepsPrivateURLs(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(sampleDDGWithPrivateURL))
	}))
	defer ts.Close()

	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	ctx := context.Background()
	results, err := Search(ctx, "test", 5)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}

	// Without P2P context, both results should be present.
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestSearch_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	origEndpoint := searchEndpoint
	defer func() { restoreEndpoint(origEndpoint) }()
	overrideEndpoint(ts.URL + "/?q=")

	_, err := Search(context.Background(), "test", 5)
	if err == nil {
		t.Fatal("expected error for HTTP 500")
	}
}

func TestClampLimit(t *testing.T) {
	tests := []struct {
		give int
		want int
	}{
		{give: 0, want: defaultLimit},
		{give: -1, want: defaultLimit},
		{give: 3, want: 3},
		{give: 5, want: 5},
		{give: 20, want: 20},
		{give: 25, want: maxLimit},
	}

	for _, tt := range tests {
		got := clampLimit(tt.give)
		if got != tt.want {
			t.Errorf("clampLimit(%d) = %d, want %d", tt.give, got, tt.want)
		}
	}
}

func TestResolveURL(t *testing.T) {
	tests := []struct {
		give string
		want string
	}{
		{
			give: "https://example.com/page",
			want: "https://example.com/page",
		},
		{
			give: "//duckduckgo.com/l/?uddg=https%3A%2F%2Freal.example.com%2Fpath&rut=abc",
			want: "https://real.example.com/path",
		},
		{
			give: "//example.com/relative",
			want: "https://example.com/relative",
		},
		{
			give: "ftp://example.com/file",
			want: "",
		},
		{
			give: "",
			want: "",
		},
	}

	for _, tt := range tests {
		got := resolveURL(tt.give)
		if got != tt.want {
			t.Errorf("resolveURL(%q) = %q, want %q", tt.give, got, tt.want)
		}
	}
}

func TestBuildTools(t *testing.T) {
	tools := BuildTools()
	if len(tools) != 1 {
		t.Fatalf("BuildTools() returned %d tools, want 1", len(tools))
	}

	tool := tools[0]
	if tool.Name != "web_search" {
		t.Errorf("Name = %q, want %q", tool.Name, "web_search")
	}
	if tool.Handler == nil {
		t.Error("Handler is nil")
	}
	if tool.Parameters == nil {
		t.Error("Parameters is nil")
	}
}
