package browser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/langoai/lango/internal/tools/browser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowserIntegration(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Setup a local test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
				<html>
					<head><title>Test Page</title></head>
					<body>
						<h1 id="header">Hello World</h1>
						<a id="more" href="/search">More Results</a>
						<button id="btn" onclick="document.getElementById('result').innerText = 'Clicked'">Click Me</button>
						<div id="result"></div>
						<input id="inp" type="text" value="">
					</body>
				</html>
			`))
		case "/search":
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
				<html>
					<head><title>Search Results</title></head>
					<body>
						<div class="result">
							<a href="https://example.com/one">Result One</a>
							<p class="result__snippet">First result snippet</p>
						</div>
						<div class="result">
							<a href="https://example.com/two">Result Two</a>
							<p class="result__snippet">Second result snippet</p>
						</div>
					</body>
				</html>
			`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// Initialize browser tool
	cfg := browser.Config{
		Headless:       true,
		SessionTimeout: 10 * time.Minute,
	}

	tool, err := browser.New(cfg)
	require.NoError(t, err, "create browser tool")
	defer tool.Close()

	// Test NewSession
	sessionID, err := tool.NewSession()
	require.NoError(t, err, "create session")

	ctx := context.Background()

	// Test Navigate
	require.NoError(t, tool.Navigate(ctx, sessionID, ts.URL), "navigate")

	// Test GetText (Header)
	text, err := tool.GetText(sessionID, "#header")
	require.NoError(t, err, "get text")
	assert.Equal(t, "Hello World", text)

	// Test Click
	require.NoError(t, tool.Click(ctx, sessionID, "#btn"), "click")

	// Wait for result update
	time.Sleep(100 * time.Millisecond)

	text, err = tool.GetText(sessionID, "#result")
	require.NoError(t, err, "get result text")
	assert.Equal(t, "Clicked", text)

	// Test Type
	require.NoError(t, tool.Type(ctx, sessionID, "#inp", "test input"), "type")

	val, err := tool.Eval(sessionID, `() => document.getElementById('inp').value`)
	require.NoError(t, err, "eval value")
	assert.Equal(t, "test input", val.(string))

	// Test Screenshot
	sst, err := tool.Screenshot(sessionID, false)
	require.NoError(t, err, "screenshot")
	assert.NotEmpty(t, sst.Data)

	// Test GetElementInfo
	info, err := tool.GetElementInfo(sessionID, "#header")
	require.NoError(t, err, "get element info")
	assert.Equal(t, "H1", info.TagName)
	assert.Equal(t, "header", info.ID)

	snapshot, err := tool.Snapshot(sessionID, 5, 5)
	require.NoError(t, err, "snapshot")
	assert.Equal(t, "generic", snapshot.PageType)
	assert.Equal(t, "Test Page", snapshot.Title)
	assert.Contains(t, snapshot.Snippet, "Hello World")
	assert.Zero(t, snapshot.ResultCount)
	assert.False(t, snapshot.Empty)
	assert.NotEmpty(t, snapshot.Links)
	assert.NotEmpty(t, snapshot.Actions)

	observed, err := tool.Observe(sessionID, 5)
	require.NoError(t, err, "observe")
	assert.NotEmpty(t, observed)

	extracted, err := tool.Extract(sessionID, "article", 5)
	require.NoError(t, err, "extract article")
	article, ok := extracted.(*browser.ArticleResult)
	require.True(t, ok)
	assert.Equal(t, "article", article.PageType)
	assert.Equal(t, ts.URL+"/", article.URL)
	assert.False(t, article.Empty)
	assert.Contains(t, article.Content, "Hello World")

	require.NoError(t, tool.Navigate(ctx, sessionID, ts.URL+"/search"), "navigate search page")
	searchSnapshot, err := tool.Snapshot(sessionID, 5, 5)
	require.NoError(t, err, "snapshot search page")
	assert.Equal(t, "search_results", searchSnapshot.PageType)
	assert.Equal(t, 2, searchSnapshot.ResultCount)
	assert.False(t, searchSnapshot.Empty)
	require.Len(t, searchSnapshot.SearchResults, 2)

	searchExtract, err := tool.Extract(sessionID, "search_results", 2)
	require.NoError(t, err, "extract search results")
	searchResults, ok := searchExtract.(*browser.SearchResponse)
	require.True(t, ok)
	assert.Equal(t, "search_results", searchResults.PageType)
	assert.Equal(t, ts.URL+"/search", searchResults.URL)
	assert.Equal(t, 2, searchResults.ResultCount)
	assert.False(t, searchResults.Empty)
	require.Len(t, searchResults.Results, 2)
	assert.Equal(t, "Result One", searchResults.Results[0].Title)
	assert.Equal(t, "https://example.com/one", searchResults.Results[0].URL)

	// Test CloseSession
	require.NoError(t, tool.CloseSession(sessionID), "close session")
}
