package sandbox

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	sandboxos "github.com/langoai/lango/internal/sandbox/os"
)

func TestNewSandboxCmdConstructsExpectedSubcommands(t *testing.T) {
	cmd := NewSandboxCmd(func() (*config.Config, error) {
		return defaultTestConfig(), nil
	}, nil)

	assert.Equal(t, "sandbox", cmd.Name())

	subcommands := make(map[string]*cobraCommandInfo)
	for _, sub := range cmd.Commands() {
		subcommands[sub.Name()] = &cobraCommandInfo{
			hidden: sub.Hidden,
			short:  sub.Short,
		}
	}

	assert.Contains(t, subcommands, "status")
	assert.Contains(t, subcommands, "test")
	require.Contains(t, subcommands, "_probe-net")
	assert.True(t, subcommands["_probe-net"].hidden)
	assert.Contains(t, subcommands["status"].short, "configuration")
	assert.Contains(t, subcommands["test"].short, "smoke tests")
}

func TestNewSandboxCmdTestBackendNoneDoesNotRunSmokeTests(t *testing.T) {
	cfgLoads := 0
	cmd := NewSandboxCmd(func() (*config.Config, error) {
		cfgLoads++
		cfg := defaultTestConfig()
		cfg.Sandbox.Enabled = true
		cfg.Sandbox.Backend = "none"
		return cfg, nil
	}, nil)
	cmd.SetArgs([]string{"test"})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	require.NoError(t, cmd.Execute())

	assert.Equal(t, 1, cfgLoads)
	assert.Contains(t, out.String(), "Sandbox backend explicitly set to none")
	assert.NotContains(t, out.String(), "Write restriction")
}

func TestNewTestCmdReturnsConfigLoaderErrorWithUsage(t *testing.T) {
	loadErr := errors.New("config unavailable")
	cmd := newTestCmd(func() (*config.Config, error) {
		return nil, loadErr
	})
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	err := cmd.Execute()

	require.ErrorIs(t, err, loadErr)
	assert.Contains(t, out.String(), "Error: config unavailable")
	assert.Contains(t, out.String(), "Usage:")
}

func TestNewTestCmdAvailableBackendRunsSmokeTestsAndReportsAllPassed(t *testing.T) {
	require.FileExists(t, "/bin/sh")

	cmd := newTestCmdWithBackendResolver(
		func() (*config.Config, error) {
			cfg := defaultTestConfig()
			cfg.Sandbox.Enabled = true
			cfg.Sandbox.Backend = "bwrap"
			return cfg, nil
		},
		func() []sandboxos.BackendCandidate { return nil },
		func(mode sandboxos.BackendMode, _ []sandboxos.BackendCandidate) (sandboxos.OSIsolator, sandboxos.BackendInfo) {
			assert.Equal(t, sandboxos.BackendBwrap, mode)
			return sandboxSmokeTestFakeIsolator{networkDenied: true}, sandboxos.BackendInfo{
				Name:      "fake",
				Mode:      sandboxos.BackendBwrap,
				Available: true,
			}
		},
	)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	require.NoError(t, cmd.Execute())

	rendered := out.String()
	assert.Contains(t, rendered, "Using isolator: fake-isolator (backend: bwrap)")
	assert.Contains(t, rendered, "Version: fake-version")
	assert.Contains(t, rendered, "Write restriction")
	assert.Contains(t, rendered, "PASS (write correctly denied)")
	assert.Contains(t, rendered, "PASS (read succeeded)")
	assert.Contains(t, rendered, "PASS (workspace write succeeded)")
	assert.Contains(t, rendered, "PASS (connect correctly denied)")
	assert.Contains(t, rendered, "All tests passed.")
}

func TestNewTestCmdAvailableBackendReportsFailedSmokeTests(t *testing.T) {
	require.FileExists(t, "/bin/sh")

	cmd := newTestCmdWithBackendResolver(
		func() (*config.Config, error) {
			cfg := defaultTestConfig()
			cfg.Sandbox.Enabled = true
			cfg.Sandbox.Backend = "bwrap"
			return cfg, nil
		},
		func() []sandboxos.BackendCandidate { return nil },
		func(sandboxos.BackendMode, []sandboxos.BackendCandidate) (sandboxos.OSIsolator, sandboxos.BackendInfo) {
			return sandboxSmokeTestFakeIsolator{networkDenied: false}, sandboxos.BackendInfo{
				Name:      "fake",
				Mode:      sandboxos.BackendBwrap,
				Available: true,
			}
		},
	)
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)

	require.NoError(t, cmd.Execute())

	rendered := out.String()
	assert.Contains(t, rendered, "FAIL (sandboxed child reached host listener)")
	assert.Contains(t, rendered, "Some tests failed.")
	assert.NotContains(t, rendered, "All tests passed.")
}

func TestNewProbeNetCmdRequiresAddressArgument(t *testing.T) {
	cmd := newProbeNetCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(nil)

	err := cmd.Execute()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts 1 arg")
	assert.Contains(t, out.String(), "Usage:")
}

type sandboxSmokeTestFakeIsolator struct {
	networkDenied bool
}

func (f sandboxSmokeTestFakeIsolator) Apply(_ context.Context, cmd *exec.Cmd, _ sandboxos.Policy) error {
	target := ""
	if len(cmd.Args) > 0 {
		target = cmd.Args[len(cmd.Args)-1]
	}

	switch {
	case target == "/etc/lango-sandbox-test":
		rewriteSandboxSmokeTestCommand(cmd, "exit 1")
	case filepath.Base(target) == "probe.txt":
		rewriteSandboxSmokeTestCommand(cmd, "touch \"$1\"", target)
	case cmd.Path == "/bin/cat":
		rewriteSandboxSmokeTestCommand(cmd, "exit 0")
	case len(cmd.Args) >= 3 && cmd.Args[1] == "sandbox" && cmd.Args[2] == "_probe-net":
		if f.networkDenied {
			rewriteSandboxSmokeTestCommand(cmd, "exit 1")
		} else {
			rewriteSandboxSmokeTestCommand(cmd, "exit 0")
		}
	default:
		rewriteSandboxSmokeTestCommand(cmd, "exit 1")
	}
	return nil
}

func (sandboxSmokeTestFakeIsolator) Available() bool { return true }
func (sandboxSmokeTestFakeIsolator) Name() string    { return "fake-isolator" }
func (sandboxSmokeTestFakeIsolator) Reason() string  { return "" }
func (sandboxSmokeTestFakeIsolator) Version() string { return "fake-version" }

func rewriteSandboxSmokeTestCommand(cmd *exec.Cmd, script string, args ...string) {
	cmd.Path = "/bin/sh"
	cmd.Args = append([]string{"sh", "-c", script, "sh"}, args...)
}

func TestReadOnlyPolicyAllowsReadsTmpWritesAndDeniesNetwork(t *testing.T) {
	policy := readOnlyPolicy()

	assert.True(t, policy.Filesystem.ReadOnlyGlobal)
	assert.Equal(t, []string{"/tmp"}, policy.Filesystem.WritePaths)
	assert.Equal(t, sandboxos.NetworkDeny, policy.Network)
	assert.True(t, policy.Process.AllowFork)
}

func TestDiscardOutputRoutesStdoutAndStderrToDiscard(t *testing.T) {
	cmd := exec.Command("ignored")

	discardOutput(cmd)

	assert.Equal(t, io.Discard, cmd.Stdout)
	assert.Equal(t, io.Discard, cmd.Stderr)
}

func TestReadTestPathSelectsPlatformReadableFile(t *testing.T) {
	got := readTestPath()

	if runtime.GOOS == "darwin" {
		assert.Equal(t, "/etc/hosts", got)
		return
	}
	assert.Equal(t, "/etc/hostname", got)
}

func TestCapabilityReasonStatusFormatsAvailabilityAndPlatformScope(t *testing.T) {
	tests := []struct {
		name             string
		available        bool
		reason           string
		currentPlatform  string
		requiredPlatform string
		want             string
	}{
		{
			name:             "available without reason",
			available:        true,
			currentPlatform:  "linux",
			requiredPlatform: "linux",
			want:             "available",
		},
		{
			name:             "available with reason",
			available:        true,
			reason:           "Landlock ABI 3",
			currentPlatform:  "linux",
			requiredPlatform: "linux",
			want:             "available (Landlock ABI 3)",
		},
		{
			name:             "different platform",
			currentPlatform:  "darwin",
			requiredPlatform: "linux",
			want:             "n/a (not on linux)",
		},
		{
			name:             "stub reason",
			reason:           "probe not yet implemented",
			currentPlatform:  "linux",
			requiredPlatform: "linux",
			want:             "unknown (probe not yet implemented)",
		},
		{
			name:             "unavailable with reason",
			reason:           "missing kernel support",
			currentPlatform:  "linux",
			requiredPlatform: "linux",
			want:             "unavailable (missing kernel support)",
		},
		{
			name:             "unavailable without reason",
			currentPlatform:  "linux",
			requiredPlatform: "linux",
			want:             "unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := capabilityReasonStatus(
				tt.available,
				tt.reason,
				tt.currentPlatform,
				tt.requiredPlatform,
			)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRenderDecisionLinesRendersGlobalAndSessionTitles(t *testing.T) {
	decisions := []statusDecision{
		{
			Timestamp:        time.Date(2026, 5, 18, 10, 11, 12, 0, time.UTC).Format("2006-01-02 15:04:05"),
			SessionKeyPrefix: "sess-123",
			Decision:         "rejected",
			Backend:          "-",
			Target:           "curl http://example.test",
			Reason:           "network disabled",
		},
	}

	var global bytes.Buffer
	renderDecisionLines(&global, decisions, "")
	assert.Contains(t, global.String(), "Recent Sandbox Decisions (global, last 10):")
	assert.Contains(t, global.String(), "[sess-123]")
	assert.Contains(t, global.String(), "curl http://example.test")
	assert.Contains(t, global.String(), "(network disabled)")

	var session bytes.Buffer
	renderDecisionLines(&session, decisions, "sess-")
	assert.Contains(t, session.String(), "Recent Sandbox Decisions (session=sess-, last 10):")
}

func TestRenderDecisionLinesSkipsEmptyDecisionSet(t *testing.T) {
	var out bytes.Buffer

	renderDecisionLines(&out, nil, "sess")

	assert.Empty(t, strings.TrimSpace(out.String()))
}

type cobraCommandInfo struct {
	hidden bool
	short  string
}
