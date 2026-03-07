package exec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	t.Parallel()

	tool := New(Config{DefaultTimeout: 5 * time.Second})

	result, err := tool.Run(context.Background(), "echo hello", 0)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.Equal(t, "hello\n", result.Stdout)
}

func TestRunTimeout(t *testing.T) {
	t.Parallel()

	tool := New(Config{DefaultTimeout: 100 * time.Millisecond})

	result, err := tool.Run(context.Background(), "sleep 10", 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, result.TimedOut, "expected timeout")
}

func TestRunWithPTY(t *testing.T) {
	t.Parallel()

	tool := New(Config{DefaultTimeout: 5 * time.Second})

	result, err := tool.RunWithPTY(context.Background(), "echo pty-test", 0)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.NotEmpty(t, result.Stdout, "expected non-empty output")
}

func TestBackgroundProcess(t *testing.T) {
	t.Parallel()

	tool := New(Config{
		DefaultTimeout:  5 * time.Second,
		AllowBackground: true,
	})
	defer tool.Cleanup()

	id, err := tool.StartBackground("sleep 10")
	require.NoError(t, err)

	status, err := tool.GetBackgroundStatus(id)
	require.NoError(t, err)
	assert.False(t, status.Done, "process should still be running")

	assert.NoError(t, tool.StopBackground(id))
}

func TestEnvFiltering(t *testing.T) {
	t.Parallel()

	tool := New(Config{})

	env := []string{
		"PATH=/usr/bin",
		"ANTHROPIC_API_KEY=secret",
		"HOME=/home/test",
	}

	filtered := tool.filterEnv(env)
	assert.Len(t, filtered, 2)

	for _, e := range filtered {
		assert.NotEqual(t, "ANTHROPIC_API_KEY=secret", e, "API key should be filtered")
	}
}

func TestFilterEnvBlacklist(t *testing.T) {
	t.Parallel()

	tool := New(Config{})

	tests := []struct {
		give     string
		wantKept bool
	}{
		{give: "PATH=/usr/bin", wantKept: true},
		{give: "HOME=/home/test", wantKept: true},
		{give: "LANGO_PASSPHRASE=supersecret", wantKept: false},
		{give: "ANTHROPIC_API_KEY=key123", wantKept: false},
		{give: "AWS_SECRET=abc", wantKept: false},
		{give: "OPENAI_API_KEY=sk-xxx", wantKept: false},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			filtered := tool.filterEnv([]string{tt.give})
			if tt.wantKept {
				assert.Len(t, filtered, 1, "expected env var to be kept")
			} else {
				assert.Empty(t, filtered, "expected env var to be filtered")
			}
		})
	}
}
