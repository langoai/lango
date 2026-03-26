package exec

import (
	"context"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	sandboxos "github.com/langoai/lango/internal/sandbox/os"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockIsolator is a test double for sandboxos.OSIsolator.
type mockIsolator struct {
	available bool
	applyErr  error
	applied   atomic.Int32
}

func (m *mockIsolator) Apply(_ context.Context, _ *exec.Cmd, _ sandboxos.Policy) error {
	m.applied.Add(1)
	return m.applyErr
}

func (m *mockIsolator) Available() bool { return m.available }
func (m *mockIsolator) Name() string    { return "mock" }

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

func TestRunSandboxIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		giveIsolator   *mockIsolator
		giveFailClosed bool
		wantApplied    int32
		wantErr        bool
		wantOutput     string
	}{
		{
			give:         "nil isolator — normal execution",
			giveIsolator: nil,
			wantApplied:  0,
			wantOutput:   "hello\n",
		},
		{
			give:         "sandbox available — Apply called, execution succeeds",
			giveIsolator: &mockIsolator{available: true},
			wantApplied:  1,
			wantOutput:   "hello\n",
		},
		{
			give:           "sandbox unavailable, fail-open — warning logged, execution continues",
			giveIsolator:   &mockIsolator{available: false, applyErr: sandboxos.ErrIsolatorUnavailable},
			giveFailClosed: false,
			wantApplied:    1,
			wantOutput:     "hello\n",
		},
		{
			give:           "sandbox unavailable, fail-closed — error returned",
			giveIsolator:   &mockIsolator{available: false, applyErr: sandboxos.ErrIsolatorUnavailable},
			giveFailClosed: true,
			wantApplied:    1,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				DefaultTimeout: 5 * time.Second,
				FailClosed:     tt.giveFailClosed,
			}
			if tt.giveIsolator != nil {
				cfg.OSIsolator = tt.giveIsolator
			}
			tool := New(cfg)

			result, err := tool.Run(context.Background(), "echo hello", 0)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
				assert.ErrorIs(t, err, sandboxos.ErrIsolatorUnavailable)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, result.Stdout)

			if tt.giveIsolator != nil {
				assert.Equal(t, tt.wantApplied, tt.giveIsolator.applied.Load())
			}
		})
	}
}

func TestRunWithPTYSandboxIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		giveIsolator   *mockIsolator
		giveFailClosed bool
		wantApplied    int32
		wantErr        bool
	}{
		{
			give:         "nil isolator — normal PTY execution",
			giveIsolator: nil,
			wantApplied:  0,
		},
		{
			give:         "sandbox available — Apply called",
			giveIsolator: &mockIsolator{available: true},
			wantApplied:  1,
		},
		{
			give:           "sandbox unavailable, fail-closed — error returned",
			giveIsolator:   &mockIsolator{available: false, applyErr: sandboxos.ErrIsolatorUnavailable},
			giveFailClosed: true,
			wantApplied:    1,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				DefaultTimeout: 5 * time.Second,
				FailClosed:     tt.giveFailClosed,
			}
			if tt.giveIsolator != nil {
				cfg.OSIsolator = tt.giveIsolator
			}
			tool := New(cfg)

			result, err := tool.RunWithPTY(context.Background(), "echo pty-test", 0)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
				assert.ErrorIs(t, err, sandboxos.ErrIsolatorUnavailable)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, result.Stdout)

			if tt.giveIsolator != nil {
				assert.Equal(t, tt.wantApplied, tt.giveIsolator.applied.Load())
			}
		})
	}
}

func TestStartBackgroundSandboxIntegration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give           string
		giveIsolator   *mockIsolator
		giveFailClosed bool
		wantApplied    int32
		wantErr        bool
	}{
		{
			give:         "nil isolator — normal background start",
			giveIsolator: nil,
			wantApplied:  0,
		},
		{
			give:         "sandbox available — Apply called",
			giveIsolator: &mockIsolator{available: true},
			wantApplied:  1,
		},
		{
			give:           "sandbox unavailable, fail-closed — error returned",
			giveIsolator:   &mockIsolator{available: false, applyErr: sandboxos.ErrIsolatorUnavailable},
			giveFailClosed: true,
			wantApplied:    1,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := Config{
				DefaultTimeout:  5 * time.Second,
				AllowBackground: true,
				FailClosed:      tt.giveFailClosed,
			}
			if tt.giveIsolator != nil {
				cfg.OSIsolator = tt.giveIsolator
			}
			tool := New(cfg)
			defer tool.Cleanup()

			id, err := tool.StartBackground("sleep 10")

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, sandboxos.ErrSandboxRequired)
				assert.ErrorIs(t, err, sandboxos.ErrIsolatorUnavailable)
				return
			}
			require.NoError(t, err)
			assert.NotEmpty(t, id)

			if tt.giveIsolator != nil {
				assert.Equal(t, tt.wantApplied, tt.giveIsolator.applied.Load())
			}

			// Clean up the background process
			require.NoError(t, tool.StopBackground(id))
		})
	}
}
