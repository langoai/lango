package exec

import (
	"context"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"

	"github.com/langoai/lango/internal/eventbus"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
	"github.com/langoai/lango/internal/security"
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
func (m *mockIsolator) Reason() string  { return "" }

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

func TestRunCapturesNonZeroExit(t *testing.T) {
	t.Parallel()

	tool := New(Config{DefaultTimeout: 5 * time.Second})

	result, err := tool.Run(context.Background(), "printf failure >&2; exit 7", 0)
	require.NoError(t, err)
	assert.Equal(t, 7, result.ExitCode)
	assert.Equal(t, "failure", result.Stderr)
}

func TestRunWithPTY(t *testing.T) {
	t.Parallel()

	tool := New(Config{DefaultTimeout: 5 * time.Second})

	result, err := tool.RunWithPTY(context.Background(), "echo pty-test", 0)
	require.NoError(t, err)
	assert.Equal(t, 0, result.ExitCode)
	assert.NotEmpty(t, result.Stdout, "expected non-empty output")
}

func TestRunWithPTYTimeoutAndNonZeroExit(t *testing.T) {
	t.Parallel()

	tool := New(Config{DefaultTimeout: 100 * time.Millisecond})

	timedOut, err := tool.RunWithPTY(context.Background(), "sleep 10", 100*time.Millisecond)
	require.NoError(t, err)
	assert.True(t, timedOut.TimedOut, "expected PTY timeout")
	assert.Equal(t, -1, timedOut.ExitCode)

	failed, err := tool.RunWithPTY(context.Background(), "exit 9", 0)
	require.NoError(t, err)
	assert.Equal(t, 9, failed.ExitCode)
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
	assert.Equal(t, id, status.ID)

	assert.NoError(t, tool.StopBackground(id))
}

func TestBackgroundProcessResidualBranches(t *testing.T) {
	t.Parallel()

	tool := New(Config{
		DefaultTimeout:  5 * time.Second,
		AllowBackground: true,
	})
	defer tool.Cleanup()

	_, err := New(Config{}).StartBackground("sleep 10")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "background processes not allowed")

	_, err = tool.GetBackgroundStatus("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "process not found: missing")

	err = tool.StopBackground("missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "process not found: missing")

	doneID, err := tool.StartBackground("printf bg-output")
	require.NoError(t, err)
	require.Eventually(t, func() bool {
		return backgroundDone(tool, doneID)
	}, 2*time.Second, 10*time.Millisecond)

	doneStatus, err := tool.GetBackgroundStatus(doneID)
	require.NoError(t, err)
	assert.Equal(t, "bg-output", doneStatus.Output.String())

	list := tool.ListBackground()
	require.Len(t, list, 1)
	assert.Equal(t, doneID, list[0].ID)

	runningID, err := tool.StartBackground("sleep 10")
	require.NoError(t, err)
	assert.NotEmpty(t, runningID)
	tool.Cleanup()
	assert.Empty(t, tool.ListBackground())
}

func backgroundDone(tool *Tool, id string) bool {
	tool.bgMu.RLock()
	defer tool.bgMu.RUnlock()

	bp, ok := tool.bgProcesses[id]
	return ok && bp.Done
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

func TestFilterEnvWhitelistAndCustomBlacklist(t *testing.T) {
	t.Parallel()

	whitelistTool := New(Config{EnvWhitelist: []string{"PATH", "HOME"}})
	filtered := whitelistTool.filterEnv([]string{
		"PATH=/usr/bin",
		"HOME=/home/test",
		"LANGO_PASSPHRASE=supersecret",
		"CUSTOM_TOKEN=secret",
	})
	assert.ElementsMatch(t, []string{"PATH=/usr/bin", "HOME=/home/test"}, filtered)

	blacklistTool := New(Config{EnvFilter: []string{"CUSTOM_TOKEN"}})
	filtered = blacklistTool.filterEnv([]string{
		"PATH=/usr/bin",
		"CUSTOM_TOKEN=secret",
		"CUSTOM_TOKEN_SUFFIX=not-filtered",
	})
	assert.ElementsMatch(t, []string{"PATH=/usr/bin", "CUSTOM_TOKEN_SUFFIX=not-filtered"}, filtered)
}

func TestResolveRefsUsesConfiguredRefStore(t *testing.T) {
	t.Parallel()

	refs := security.NewRefStore()
	token := refs.Store("api-key", []byte("resolved-secret"))
	tool := New(Config{Refs: refs})

	assert.Equal(t, "echo resolved-secret", tool.resolveRefs("echo "+token))
}

func TestSetEventBusAndSyncBuffer(t *testing.T) {
	t.Parallel()

	bus := eventbus.New()
	tool := New(Config{})
	tool.SetEventBus(bus)
	assert.Same(t, bus, tool.config.Bus)

	var sb syncBuffer
	n, err := sb.Write([]byte("hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, "hello", sb.String())
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

// TestExcludedMatch verifies the basename matcher used by ExcludedCommands.
// IMPORTANT: this matcher consumes the user command string (NOT cmd.Args[0])
// because exec.Tool wraps every invocation in `sh -c <command>`. The first
// token of the user command is the actual program name; cmd.Args[0] is "sh".
func TestExcludedMatch(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give        string
		givePats    []string
		wantMatched string
		wantPattern string
	}{
		{give: "git status", givePats: []string{"git"}, wantMatched: "git", wantPattern: "git"},
		{give: "/usr/bin/git push", givePats: []string{"git"}, wantMatched: "git", wantPattern: "git"},
		{give: "git status | grep foo", givePats: []string{"git"}, wantMatched: "git", wantPattern: "git"},
		// Conservative: chained commands match the FIRST token only.
		{give: "cd /tmp && git status", givePats: []string{"git"}, wantMatched: "", wantPattern: ""},
		// Empty inputs.
		{give: "", givePats: []string{"git"}, wantMatched: "", wantPattern: ""},
		{give: "git status", givePats: nil, wantMatched: "", wantPattern: ""},
		// No match.
		{give: "echo hi", givePats: []string{"git", "docker"}, wantMatched: "", wantPattern: ""},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			matched, pattern := excludedMatch(tt.give, tt.givePats)
			assert.Equal(t, tt.wantMatched, matched)
			assert.Equal(t, tt.wantPattern, pattern)
		})
	}
}

// TestApplySandbox_ExcludedDoesNotMatchSh is a regression guard against the
// "cmd.Args[0] is 'sh'" trap. If a future refactor accidentally matches on
// cmd.Args[0] instead of the user command, every command would bypass the
// sandbox whenever excluded=["sh"]. This test pins the correct semantics:
// the sh wrapper is invisible to ExcludedCommands matching.
func TestApplySandbox_ExcludedDoesNotMatchSh(t *testing.T) {
	t.Parallel()

	iso := &mockIsolator{available: true}
	tool := New(Config{
		DefaultTimeout:   5 * time.Second,
		OSIsolator:       iso,
		ExcludedCommands: []string{"sh"},
	})

	_, err := tool.Run(context.Background(), "echo hello", 0)
	require.NoError(t, err)
	assert.Equal(t, int32(1), iso.applied.Load(),
		"isolator must be applied; ExcludedCommands=['sh'] must NOT bypass sandbox just because the wrapper command is sh -c")
}

// TestApplySandbox_ExcludedBypass verifies that a matching basename causes
// applySandbox to skip the isolator entirely.
func TestApplySandbox_ExcludedBypass(t *testing.T) {
	t.Parallel()

	iso := &mockIsolator{available: true}
	tool := New(Config{
		DefaultTimeout:   5 * time.Second,
		OSIsolator:       iso,
		ExcludedCommands: []string{"echo"},
	})

	_, err := tool.Run(context.Background(), "echo hello", 0)
	require.NoError(t, err)
	assert.Equal(t, int32(0), iso.applied.Load(),
		"isolator must NOT be applied for an excluded command")
}
