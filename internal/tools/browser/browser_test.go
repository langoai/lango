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
						<button id="btn" onclick="document.getElementById('result').innerText = 'Clicked'">Click Me</button>
						<div id="result"></div>
						<input id="inp" type="text" value="">
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

	// Test CloseSession
	require.NoError(t, tool.CloseSession(sessionID), "close session")
}
