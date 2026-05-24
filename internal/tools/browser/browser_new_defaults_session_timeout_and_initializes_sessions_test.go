package browser

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBrowserNewDefaultsSessionTimeoutAndInitializesSessions(t *testing.T) {
	t.Parallel()

	tool, err := New(Config{})
	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, tool.config.SessionTimeout)
	assert.NotNil(t, tool.sessions)
	assert.False(t, tool.initDone)

	explicitTimeout := 30 * time.Second
	tool, err = New(Config{SessionTimeout: explicitTimeout})
	require.NoError(t, err)
	assert.Equal(t, explicitTimeout, tool.config.SessionTimeout)
}

func TestBrowserSessionMethodsReturnNotFoundWithoutStartingBrowser(t *testing.T) {
	t.Parallel()

	tool := &Tool{sessions: make(map[string]*Session)}
	ctx := context.Background()

	_, err := tool.getSession("missing")
	require.Error(t, err)
	assert.EqualError(t, err, "session not found: missing")
	assert.False(t, tool.HasSession("missing"))

	checks := []struct {
		name string
		run  func() error
	}{
		{name: "navigate", run: func() error { return tool.Navigate(ctx, "missing", "https://example.test") }},
		{name: "current url", run: func() error { _, err := tool.CurrentURL("missing"); return err }},
		{name: "screenshot", run: func() error { _, err := tool.Screenshot("missing", false); return err }},
		{name: "click", run: func() error { return tool.Click(ctx, "missing", "#button") }},
		{name: "type", run: func() error { return tool.Type(ctx, "missing", "#input", "hello") }},
		{name: "get text", run: func() error { _, err := tool.GetText("missing", "#title"); return err }},
		{name: "snapshot", run: func() error { _, err := tool.GetSnapshot("missing"); return err }},
		{name: "element info", run: func() error { _, err := tool.GetElementInfo("missing", "#title"); return err }},
		{name: "eval", run: func() error { _, err := tool.Eval("missing", "() => 1"); return err }},
		{name: "wait", run: func() error { return tool.WaitForSelector(ctx, "missing", "#ready", 0) }},
		{name: "close", run: func() error { return tool.CloseSession("missing") }},
	}

	for _, check := range checks {
		check := check
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()

			err := check.run()
			require.Error(t, err)
			assert.EqualError(t, err, "session not found: missing")
		})
	}
}

func TestBrowserNilPageOperationsRecoverRodPanic(t *testing.T) {
	t.Parallel()

	tool := &Tool{
		sessions: map[string]*Session{
			"nil-page": {ID: "nil-page"},
		},
	}
	ctx := context.Background()

	checks := []struct {
		name      string
		run       func() error
		wantError string
	}{
		{
			name:      "navigate",
			run:       func() error { return tool.Navigate(ctx, "nil-page", "https://example.test") },
			wantError: "browser panic recovered",
		},
		{
			name: "current url",
			run: func() error {
				_, err := tool.CurrentURL("nil-page")
				return err
			},
			wantError: "get current URL: browser panic recovered",
		},
		{
			name: "screenshot",
			run: func() error {
				_, err := tool.Screenshot("nil-page", false)
				return err
			},
			wantError: "screenshot: browser panic recovered",
		},
		{
			name:      "click",
			run:       func() error { return tool.Click(ctx, "nil-page", "#button") },
			wantError: "browser panic recovered",
		},
		{
			name:      "type",
			run:       func() error { return tool.Type(ctx, "nil-page", "#input", "hello") },
			wantError: "browser panic recovered",
		},
		{
			name: "get text",
			run: func() error {
				_, err := tool.GetText("nil-page", "#title")
				return err
			},
			wantError: "browser panic recovered",
		},
		{
			name: "snapshot",
			run: func() error {
				_, err := tool.GetSnapshot("nil-page")
				return err
			},
			wantError: "browser panic recovered",
		},
		{
			name: "element info",
			run: func() error {
				_, err := tool.GetElementInfo("nil-page", "#title")
				return err
			},
			wantError: "browser panic recovered",
		},
		{
			name: "eval",
			run: func() error {
				_, err := tool.Eval("nil-page", "() => 1")
				return err
			},
			wantError: "browser panic recovered",
		},
		{
			name:      "wait custom timeout",
			run:       func() error { return tool.WaitForSelector(ctx, "nil-page", "#ready", time.Millisecond) },
			wantError: "browser panic recovered",
		},
	}

	for _, check := range checks {
		check := check
		t.Run(check.name, func(t *testing.T) {
			t.Parallel()

			err := check.run()
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrBrowserPanic)
			assert.Contains(t, err.Error(), check.wantError)
		})
	}
}
