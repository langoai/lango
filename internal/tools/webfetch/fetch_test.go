package webfetch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/langoai/lango/internal/agent"
	"github.com/langoai/lango/internal/ctxkeys"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const sampleHTML = `<!DOCTYPE html>
<html>
<head><title>Test Page</title></head>
<body>
<nav>Navigation bar</nav>
<main>
<h1>Hello World</h1>
<p>This is a <strong>test</strong> paragraph with some content.</p>
<ul>
<li>Item one</li>
<li>Item two</li>
</ul>
<a href="https://example.com">Example Link</a>
</main>
<footer>Footer content</footer>
<script>var x = 1;</script>
</body>
</html>`

func newTestServer(statusCode int, body string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(body))
	}))
}

func TestFetch_TextMode(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	result, err := Fetch(context.Background(), srv.URL, ModeText, defaultMaxLength, false)
	require.NoError(t, err)

	assert.Equal(t, srv.URL, result.URL)
	assert.Equal(t, "Test Page", result.Title)
	assert.Contains(t, result.Content, "Hello World")
	assert.Contains(t, result.Content, "test")
	assert.Contains(t, result.Content, "paragraph")
	// Nav and footer should be stripped.
	assert.NotContains(t, result.Content, "Navigation bar")
	assert.NotContains(t, result.Content, "Footer content")
	// Script should be stripped.
	assert.NotContains(t, result.Content, "var x = 1")
	assert.False(t, result.Truncated)
}

func TestFetch_HTMLMode(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	result, err := Fetch(context.Background(), srv.URL, ModeHTML, defaultMaxLength, false)
	require.NoError(t, err)

	assert.Equal(t, "Test Page", result.Title)
	// HTML mode returns raw HTML.
	assert.Contains(t, result.Content, "<h1>Hello World</h1>")
	assert.Contains(t, result.Content, "<nav>")
}

func TestFetch_MarkdownMode(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	result, err := Fetch(context.Background(), srv.URL, ModeMarkdown, defaultMaxLength, false)
	require.NoError(t, err)

	assert.Equal(t, "Test Page", result.Title)
	// Headings should be markdown.
	assert.Contains(t, result.Content, "# Hello World")
	// Links should be markdown format.
	assert.Contains(t, result.Content, "[Example Link](https://example.com)")
	// List items should have dash prefix.
	assert.Contains(t, result.Content, "- Item one")
	// Nav and footer should be stripped.
	assert.NotContains(t, result.Content, "Navigation bar")
	assert.NotContains(t, result.Content, "Footer content")
}

func TestFetch_MaxLengthTruncation(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	result, err := Fetch(context.Background(), srv.URL, ModeText, 20, false)
	require.NoError(t, err)

	assert.True(t, result.Truncated)
	assert.Equal(t, 20, result.ContentLength)
	assert.Len(t, result.Content, 20)
}

func TestFetch_Non200Error(t *testing.T) {
	srv := newTestServer(http.StatusNotFound, "not found")
	defer srv.Close()

	_, err := Fetch(context.Background(), srv.URL, ModeText, defaultMaxLength, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
}

func TestFetch_ServerError(t *testing.T) {
	srv := newTestServer(http.StatusInternalServerError, "internal error")
	defer srv.Close()

	_, err := Fetch(context.Background(), srv.URL, ModeText, defaultMaxLength, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 500")
}

func TestFetch_InvalidMode(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	_, err := Fetch(context.Background(), srv.URL, "xml", defaultMaxLength, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported mode")
}

func TestFetch_EmptyURL(t *testing.T) {
	_, err := Fetch(context.Background(), "", ModeText, defaultMaxLength, false)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty URL")
}

func TestFetch_DefaultMode(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	// Empty mode defaults to text.
	result, err := Fetch(context.Background(), srv.URL, "", defaultMaxLength, false)
	require.NoError(t, err)
	assert.Contains(t, result.Content, "Hello World")
}

func TestFetch_DefaultMaxLength(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	// Zero maxLength defaults to defaultMaxLength.
	result, err := Fetch(context.Background(), srv.URL, ModeText, 0, false)
	require.NoError(t, err)
	assert.NotEmpty(t, result.Content)
}

func TestFetch_URLWithoutScheme(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	// The test server URL has http:// prefix. Strip it and pass without scheme.
	// Fetch prepends https:// which won't match the http test server, causing a TLS error.
	noScheme := strings.TrimPrefix(srv.URL, "http://")
	_, err := Fetch(context.Background(), noScheme, ModeText, defaultMaxLength, false)
	// Should get an error because https:// is prepended but server speaks plain http.
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "parse URL")
}

func TestFetch_Redirect(t *testing.T) {
	final := newTestServer(http.StatusOK, sampleHTML)
	defer final.Close()

	redirect := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, final.URL, http.StatusMovedPermanently)
	}))
	defer redirect.Close()

	result, err := Fetch(context.Background(), redirect.URL, ModeText, defaultMaxLength, false)
	require.NoError(t, err)
	assert.Equal(t, final.URL, result.URL)
	assert.Contains(t, result.Content, "Hello World")
}

func TestValidateURLForP2P(t *testing.T) {
	tests := []struct {
		give    string
		wantErr bool
	}{
		{give: "https://example.com", wantErr: false},
		{give: "file:///etc/passwd", wantErr: true},
		{give: "http://localhost:8080", wantErr: true},
		{give: "http://127.0.0.1:8080", wantErr: true},
		{give: "http://[::1]:8080", wantErr: true},
		{give: "http://10.0.0.1/secret", wantErr: true},
		{give: "http://192.168.1.1/admin", wantErr: true},
		{give: "http://172.16.0.1/internal", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := ValidateURLForP2P(tt.give)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestFetch_P2PContextBlocksPrivateURL(t *testing.T) {
	ctx := ctxkeys.WithP2PRequest(context.Background())

	// Build a test server (binds to 127.0.0.1 which is private).
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	// The Fetch function itself doesn't check P2P — that's the tool handler's job.
	// But ValidateURLForP2P should block it.
	err := ValidateURLForP2P(srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "loopback")

	// Verify context carries P2P flag.
	assert.True(t, ctxkeys.IsP2PRequest(ctx))
}

func TestExtractText_Readability(t *testing.T) {
	tests := []struct {
		give       string
		wantTitle  string
		wantIn     []string
		wantNotIn  []string
	}{
		{
			give: `<html><head><title>T</title></head><body>
				<nav>Skip</nav>
				<article><h1>Title</h1><p>Content here</p></article>
				<footer>Skip too</footer>
			</body></html>`,
			wantTitle: "T",
			wantIn:    []string{"Title", "Content here"},
			wantNotIn: []string{"Skip"},
		},
		{
			give: `<html><head><title>Main</title></head><body>
				<main><p>Main content</p></main>
				<aside>Sidebar</aside>
			</body></html>`,
			wantTitle: "Main",
			wantIn:    []string{"Main content"},
			wantNotIn: []string{"Sidebar"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantTitle, func(t *testing.T) {
			title, body, err := extractText(strings.NewReader(tt.give))
			require.NoError(t, err)
			assert.Equal(t, tt.wantTitle, title)
			for _, want := range tt.wantIn {
				assert.Contains(t, body, want)
			}
			for _, notWant := range tt.wantNotIn {
				assert.NotContains(t, body, notWant)
			}
		})
	}
}

func TestExtractMarkdown(t *testing.T) {
	input := `<html><head><title>MD</title></head><body>
		<main>
			<h1>Heading One</h1>
			<h2>Heading Two</h2>
			<p>Para with <strong>bold</strong> and <em>italic</em></p>
			<a href="https://go.dev">Go</a>
			<ul><li>First</li><li>Second</li></ul>
			<pre>code block</pre>
		</main>
	</body></html>`

	title, body, err := extractMarkdown(strings.NewReader(input))
	require.NoError(t, err)

	assert.Equal(t, "MD", title)
	assert.Contains(t, body, "# Heading One")
	assert.Contains(t, body, "## Heading Two")
	assert.Contains(t, body, "**bold**")
	assert.Contains(t, body, "*italic*")
	assert.Contains(t, body, "[Go](https://go.dev)")
	assert.Contains(t, body, "- First")
	assert.Contains(t, body, "- Second")
	assert.Contains(t, body, "```\ncode block\n```")
}

func TestBuildTools_ToolDefinition(t *testing.T) {
	tools := BuildTools()
	require.Len(t, tools, 1)

	tool := tools[0]
	assert.Equal(t, "web_fetch", tool.Name)
	assert.Equal(t, "Fetch a web page and extract its content. Supports text, HTML, and markdown output modes.", tool.Description)
	assert.Equal(t, agent.SafetyLevelModerate, tool.SafetyLevel)
	assert.Equal(t, "web", tool.Capability.Category)
	assert.Equal(t, agent.ActivityRead, tool.Capability.Activity)
	assert.Contains(t, tool.Capability.Aliases, "fetch_url")
	assert.Contains(t, tool.Capability.Aliases, "get_page")
	assert.NotNil(t, tool.Handler)
}

func TestBuildTools_Handler(t *testing.T) {
	srv := newTestServer(http.StatusOK, sampleHTML)
	defer srv.Close()

	tools := BuildTools()
	handler := tools[0].Handler

	result, err := handler(context.Background(), map[string]interface{}{
		"url":  srv.URL,
		"mode": ModeText,
	})
	require.NoError(t, err)

	fr, ok := result.(*FetchResult)
	require.True(t, ok)
	assert.Equal(t, "Test Page", fr.Title)
	assert.Contains(t, fr.Content, "Hello World")
}

func TestBuildTools_HandlerMissingURL(t *testing.T) {
	tools := BuildTools()
	handler := tools[0].Handler

	_, err := handler(context.Background(), map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing url")
}
