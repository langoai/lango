package browser_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/langoai/lango/internal/tools/browser"
)

func TestBrowserIntegration(t *testing.T) {
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
	if err != nil {
		t.Fatalf("failed to create browser tool: %v", err)
	}
	defer tool.Close()

	// Test NewSession
	sessionID, err := tool.NewSession()
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	ctx := context.Background()

	// Test Navigate
	if err := tool.Navigate(ctx, sessionID, ts.URL); err != nil {
		t.Fatalf("failed to navigate: %v", err)
	}

	// Test GetText (Header)
	text, err := tool.GetText(sessionID, "#header")
	if err != nil {
		t.Fatalf("failed to get text: %v", err)
	}
	if text != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", text)
	}

	// Test Click
	if err := tool.Click(ctx, sessionID, "#btn"); err != nil {
		t.Fatalf("failed to click: %v", err)
	}

	// Wait for result update
	time.Sleep(100 * time.Millisecond) // simple wait, ideally use WaitForSelector logic or Eval

	text, err = tool.GetText(sessionID, "#result")
	if err != nil {
		t.Fatalf("failed to get result text: %v", err)
	}
	if text != "Clicked" {
		t.Errorf("expected 'Clicked', got '%s'", text)
	}

	// Test Type
	if err := tool.Type(ctx, sessionID, "#inp", "test input"); err != nil {
		t.Fatalf("failed to type: %v", err)
	}

	val, err := tool.Eval(sessionID, `() => document.getElementById('inp').value`)
	if err != nil {
		t.Fatalf("failed to eval value: %v", err)
	}
	if val.(string) != "test input" {
		t.Errorf("expected 'test input', got '%s'", val)
	}

	// Test Screenshot
	sst, err := tool.Screenshot(sessionID, false)
	if err != nil {
		t.Fatalf("failed to screenshot: %v", err)
	}
	if len(sst.Data) == 0 {
		t.Error("screenshot data empty")
	}

	// Test GetElementInfo
	info, err := tool.GetElementInfo(sessionID, "#header")
	if err != nil {
		t.Fatalf("failed to get element info: %v", err)
	}
	if info.TagName != "H1" {
		t.Errorf("expected tag H1, got %s", info.TagName)
	}
	if info.ID != "header" {
		t.Errorf("expected id header, got %s", info.ID)
	}

	// Test CloseSession
	if err := tool.CloseSession(sessionID); err != nil {
		t.Fatalf("failed to close session: %v", err)
	}
}
