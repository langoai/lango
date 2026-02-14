package exec

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	tool := New(Config{DefaultTimeout: 5 * time.Second})

	result, err := tool.Run(context.Background(), "echo hello", 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if result.Stdout != "hello\n" {
		t.Errorf("expected 'hello\\n', got %q", result.Stdout)
	}
}

func TestRunTimeout(t *testing.T) {
	tool := New(Config{DefaultTimeout: 100 * time.Millisecond})

	result, err := tool.Run(context.Background(), "sleep 10", 100*time.Millisecond)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if !result.TimedOut {
		t.Error("expected timeout")
	}
}

func TestRunWithPTY(t *testing.T) {
	tool := New(Config{DefaultTimeout: 5 * time.Second})

	result, err := tool.RunWithPTY(context.Background(), "echo pty-test", 0)
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	// PTY output includes the echoed command
	if len(result.Stdout) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestBackgroundProcess(t *testing.T) {
	tool := New(Config{
		DefaultTimeout:  5 * time.Second,
		AllowBackground: true,
	})
	defer tool.Cleanup()

	id, err := tool.StartBackground("sleep 10")
	if err != nil {
		t.Fatalf("failed to start: %v", err)
	}

	status, err := tool.GetBackgroundStatus(id)
	if err != nil {
		t.Fatalf("failed to get status: %v", err)
	}

	if status.Done {
		t.Error("process should still be running")
	}

	if err := tool.StopBackground(id); err != nil {
		t.Errorf("failed to stop: %v", err)
	}
}

func TestEnvFiltering(t *testing.T) {
	tool := New(Config{})

	env := []string{
		"PATH=/usr/bin",
		"ANTHROPIC_API_KEY=secret",
		"HOME=/home/test",
	}

	filtered := tool.filterEnv(env)
	if len(filtered) != 2 {
		t.Errorf("expected 2 vars, got %d", len(filtered))
	}

	for _, e := range filtered {
		if e == "ANTHROPIC_API_KEY=secret" {
			t.Error("API key should be filtered")
		}
	}
}

func TestFilterEnvBlacklist(t *testing.T) {
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
			filtered := tool.filterEnv([]string{tt.give})
			if tt.wantKept {
				assert.Len(t, filtered, 1, "expected env var to be kept")
			} else {
				assert.Empty(t, filtered, "expected env var to be filtered")
			}
		})
	}
}
