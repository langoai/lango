package sandbox

import (
	"bytes"
	"context"
	"errors"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
	"github.com/langoai/lango/internal/storage"
)

type collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator struct {
	applyErr error
	rewrite  func(*exec.Cmd)
}

func (f collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator) Apply(_ context.Context, cmd *exec.Cmd, _ sandboxos.Policy) error {
	if f.applyErr != nil {
		return f.applyErr
	}
	if f.rewrite != nil {
		f.rewrite(cmd)
	}
	return nil
}

func (f collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator) Available() bool {
	return true
}
func (f collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator) Name() string {
	return "runChatValidModeReachesAppBuilderBeforeTui2-fake"
}
func (f collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator) Reason() string {
	return ""
}

func collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t *testing.T, code int) func(*exec.Cmd) {
	t.Helper()
	require.FileExists(t, "/bin/sh")
	return func(cmd *exec.Cmd) {
		cmd.Path = "/bin/sh"
		cmd.Args = []string{"sh", "-c", "exit " + strconv.Itoa(code)}
	}
}

func TestCollectRecentDecisionsConvertsRecordsAndSwallowsErrors(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 19, 9, 10, 11, 0, time.UTC)
	var gotPrefix string
	var gotLimit int
	boot := &bootstrap.Result{
		Storage: storage.NewFacade(nil, nil, storage.WithSandboxDecisionReader(
			func(_ context.Context, sessionPrefix string, limit int) ([]storage.SandboxDecisionRecord, error) {
				gotPrefix = sessionPrefix
				gotLimit = limit
				return []storage.SandboxDecisionRecord{
					{
						Timestamp:  now,
						SessionKey: "session-abcdef",
						Decision:   "applied",
						Backend:    "bwrap",
						Target:     "go test ./...",
					},
					{
						Timestamp:  now.Add(time.Second),
						SessionKey: "short",
						Decision:   "rejected",
						Backend:    "seatbelt",
						Target:     "curl http://example.test",
						Reason:     "network disabled",
					},
				}, nil
			},
		)),
	}

	decisions := collectRecentDecisions(context.Background(), boot, "session-")

	require.Len(t, decisions, 2)
	assert.Equal(t, "session-", gotPrefix)
	assert.Equal(t, 10, gotLimit)
	assert.Equal(t, "2026-05-19 09:10:11", decisions[0].Timestamp)
	assert.Equal(t, "session-", decisions[0].SessionKeyPrefix)
	assert.Equal(t, "bwrap", decisions[0].Backend)
	assert.Equal(t, "-", decisions[1].Backend)
	assert.Equal(t, "network disabled", decisions[1].Reason)

	failingBoot := &bootstrap.Result{
		Storage: storage.NewFacade(nil, nil, storage.WithSandboxDecisionReader(
			func(context.Context, string, int) ([]storage.SandboxDecisionRecord, error) {
				return nil, errors.New("audit unavailable")
			},
		)),
	}
	assert.Empty(t, collectRecentDecisions(context.Background(), failingBoot, "session-"))
}

func TestBuildStatusSnapshotCoversOptOutFailModesAndWarnings(t *testing.T) {
	t.Parallel()

	cfg := defaultTestConfig()
	cfg.Sandbox.Enabled = true
	cfg.Sandbox.FailClosed = true
	cfg.Sandbox.Backend = "none"
	cfg.Sandbox.WorkspacePath = ""
	cfg.Sandbox.NetworkMode = "deny"
	cfg.Sandbox.AllowedNetworkIPs = []string{"127.0.0.1"}

	snapshot := buildStatusSnapshot(context.Background(), cfg, nil, "sess")

	assert.True(t, snapshot.Configuration.Enabled)
	assert.True(t, snapshot.Configuration.ExplicitOptOut)
	assert.Equal(t, "none", snapshot.Configuration.Backend)
	assert.Equal(t, "none (explicit opt-out - fail-closed not applied)", snapshot.Configuration.BackendLabel)
	assert.Empty(t, snapshot.Configuration.FailMode)
	assert.NotEmpty(t, snapshot.Configuration.Workspace)
	assert.Equal(t, "disabled", snapshot.ActiveIsolation.Isolator)
	assert.False(t, snapshot.ActiveIsolation.Available)
	if runtimeGOOSForCollectRecentDecisionsConvertsRecordsAndSwallowsErrors() == "linux" {
		assert.True(t, snapshot.Warnings.AllowedNetworkIPsMacOSOnly)
	}

	cfg.Sandbox.Backend = "native"
	snapshot = buildStatusSnapshot(context.Background(), cfg, nil, "")
	assert.Equal(t, "fail-closed (execution rejected)", snapshot.Configuration.FailMode)
	assert.Equal(t, "native", snapshot.Configuration.BackendLabel)
	assert.Equal(t, "native backend not yet implemented", snapshot.ActiveIsolation.Reason)
}

func TestRenderStatusTableIncludesUnavailableAndWarningBranches(t *testing.T) {
	t.Parallel()

	snapshot := statusSnapshot{
		SessionPrefix: "sess-",
		Configuration: statusConfiguration{
			Enabled:        true,
			FailMode:       "fail-open (warning + unsandboxed execution)",
			BackendLabel:   "auto (resolved: bwrap)",
			NetworkMode:    "deny",
			Workspace:      "/tmp/lango-runChatValidModeReachesAppBuilderBeforeTui2",
			ExplicitOptOut: false,
		},
		ActiveIsolation: statusActiveIsolation{
			Isolator:                    "bwrap",
			Available:                   true,
			NetworkIsolationUnavailable: true,
			NetworkIsolationReason:      "unshare-net unavailable",
		},
		PlatformCapabilities: statusPlatformCapabilities{
			Platform: "linux",
			Kernel:   "6.test",
			Seatbelt: "n/a (not on darwin)",
			Landlock: "available",
			Seccomp:  "unavailable",
		},
		BackendAvailability: []statusBackend{
			{Name: "bwrap", Available: true},
			{Name: "native", Available: false, Reason: "native backend not yet implemented"},
		},
		RecentDecisions: []statusDecision{
			{
				Timestamp:        "2026-05-19 09:10:11",
				SessionKeyPrefix: "sess-123",
				Decision:         "rejected",
				Backend:          "-",
				Target:           "curl http://example.test",
				Reason:           "network disabled",
			},
		},
		Warnings: statusWarnings{AllowedNetworkIPsMacOSOnly: true},
	}

	var out bytes.Buffer
	renderStatusTable(&out, snapshot)

	rendered := out.String()
	assert.Contains(t, rendered, "Fail-Closed:    fail-open")
	assert.Contains(t, rendered, "Network Iso:    unavailable (unshare-net unavailable)")
	assert.Contains(t, rendered, "native:         unavailable (native backend not yet implemented)")
	assert.Contains(t, rendered, "WARNING: allowedNetworkIPs is macOS-only")
	assert.Contains(t, rendered, "Recent Sandbox Decisions (session=sess-, last 10):")
	assert.Contains(t, rendered, "curl http://example.test")
}

func TestStatusCommandValidationAndLoaderErrors(t *testing.T) {
	t.Parallel()

	t.Run("invalid output", func(t *testing.T) {
		cmd := newStatusCmd(func() (*config.Config, error) {
			return defaultTestConfig(), nil
		}, nil)
		cmd.SetArgs([]string{"--output", "yaml"})
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported output format")
	})

	t.Run("missing loaders", func(t *testing.T) {
		cmd := newStatusCmd(nil, nil)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "requires a config loader or bootstrap loader")
	})

	t.Run("config loader error", func(t *testing.T) {
		wantErr := errors.New("config load failed")
		cmd := newStatusCmd(func() (*config.Config, error) {
			return nil, wantErr
		}, nil)
		var out bytes.Buffer
		cmd.SetOut(&out)
		cmd.SetErr(&out)

		err := cmd.Execute()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "load config")
		assert.ErrorIs(t, err, wantErr)
	})
}

func TestNewTestCmdNativeBackendUnavailableSkipsSmokeTests(t *testing.T) {
	t.Parallel()

	cmd := newTestCmd(func() (*config.Config, error) {
		cfg := defaultTestConfig()
		cfg.Sandbox.Enabled = true
		cfg.Sandbox.Backend = "native"
		return cfg, nil
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	require.NoError(t, cmd.Execute())

	rendered := out.String()
	assert.Contains(t, rendered, "Sandbox backend native not available")
	assert.Contains(t, rendered, "native backend not yet implemented")
	assert.NotContains(t, rendered, "Write restriction")
}

func TestProbeNetCommandConnectsToLoopbackAndReportsDialFailure(t *testing.T) {
	t.Parallel()

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	accepted := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err == nil {
			_ = conn.Close()
			close(accepted)
		}
	}()

	cmd := newProbeNetCmd()
	cmd.SetArgs([]string{ln.Addr().String()})
	require.NoError(t, cmd.Execute())

	select {
	case <-accepted:
	case <-time.After(2 * time.Second):
		t.Fatal("probe command did not connect to the loopback listener")
	}

	cmd = newProbeNetCmd()
	cmd.SetArgs([]string{"127.0.0.1:0"})
	err = cmd.Execute()
	require.Error(t, err)
	assert.NotContains(t, strings.ToLower(err.Error()), "usage")
}

func TestRunSmokeTestHelpersUseApplyAndCommandOutcome(t *testing.T) {
	t.Parallel()

	applyErr := errors.New("apply failed")
	assert.False(t, runReadTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{applyErr: applyErr}))

	assert.True(t, runReadTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{rewrite: collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t, 0)}))
	assert.False(t, runReadTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{rewrite: collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t, 1)}))

	assert.True(t, runWriteTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{rewrite: collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t, 1)}))
	assert.False(t, runWriteTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{rewrite: collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t, 0)}))

	assert.True(t, runNetworkDenyTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{rewrite: collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t, 1)}))
	assert.False(t, runNetworkDenyTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{rewrite: collectRecentDecisionsConvertsRecordsAndSwallowsErrorsRewriteShellExit(t, 0)}))
	assert.False(t, runNetworkDenyTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{applyErr: applyErr}))
}

func TestRunWorkspaceWriteTestCreatesAllowedFile(t *testing.T) {
	t.Parallel()

	assert.True(t, runWorkspaceWriteTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{}))
	assert.False(t, runWorkspaceWriteTest(collectRecentDecisionsConvertsRecordsAndSwallowsErrorsFakeIsolator{applyErr: errors.New("apply failed")}))
}

func TestFindTouchFallsBackWhenPathLookupMisses(t *testing.T) {
	var want string
	for _, candidate := range []string{"/usr/bin/touch", "/bin/touch"} {
		fi, err := os.Stat(candidate)
		if err == nil && !fi.IsDir() {
			want = candidate
			break
		}
	}
	if want == "" {
		t.Skip("platform has no standard touch fallback path")
	}

	emptyPath := t.TempDir()
	t.Setenv("PATH", emptyPath)

	got := findTouch()

	assert.Equal(t, want, got)
	assert.Equal(t, "touch", filepath.Base(got))
}

func runtimeGOOSForCollectRecentDecisionsConvertsRecordsAndSwallowsErrors() string {
	return runtime.GOOS
}
