package p2p

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/sandbox"
	"github.com/langoai/lango/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubNamedSandboxExecutor struct {
	runtimeName string
	result      map[string]interface{}
	err         error
}

func (s stubNamedSandboxExecutor) Execute(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
	return s.result, s.err
}

func (s stubNamedSandboxExecutor) RuntimeName() string {
	return s.runtimeName
}

type stubSandboxExecutor struct {
	result map[string]interface{}
	err    error
}

func (s stubSandboxExecutor) Execute(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
	return s.result, s.err
}

type stubCleanupRuntime struct {
	available  bool
	cleanupErr error
}

func (s stubCleanupRuntime) IsAvailable(_ context.Context) bool {
	return s.available
}

func (s stubCleanupRuntime) Cleanup(_ context.Context, _ string) error {
	return s.cleanupErr
}

func TestSandboxStatusCmd_WritesDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.ToolIsolation.Enabled = false

	cmd := newSandboxStatusCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Tool isolation: disabled")
}

func TestSandboxStatusCmd_WritesContainerDisabledStateToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.P2P.ToolIsolation.Enabled = true
	cfg.P2P.ToolIsolation.TimeoutPerTool = 30 * time.Second
	cfg.P2P.ToolIsolation.MaxMemoryMB = 512
	cfg.P2P.ToolIsolation.Container.Enabled = false

	cmd := newSandboxStatusCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Tool isolation: enabled")
	assert.Contains(t, out, "Container mode:   disabled (subprocess fallback)")
}

func TestSandboxStatusCmd_WritesUnavailableRuntimeToCommandWriter(t *testing.T) {
	original := newContainerSandboxExecutor
	newContainerSandboxExecutor = func(cfg sandbox.Config, containerCfg config.ContainerSandboxConfig) (namedSandboxExecutor, error) {
		return nil, errors.New("runtime probe failed")
	}
	t.Cleanup(func() { newContainerSandboxExecutor = original })

	cfg := config.DefaultConfig()
	cfg.P2P.ToolIsolation.Enabled = true
	cfg.P2P.ToolIsolation.Container.Enabled = true
	cfg.P2P.ToolIsolation.Container.Runtime = "auto"
	cfg.P2P.ToolIsolation.Container.Image = "lango-sandbox:latest"
	cfg.P2P.ToolIsolation.Container.NetworkMode = "none"

	cmd := newSandboxStatusCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Active runtime:   unavailable (runtime probe failed)")
}

func TestSandboxTestCmd_WritesSubprocessPathToCommandWriter(t *testing.T) {
	original := newSubprocessSandboxExecutor
	newSubprocessSandboxExecutor = func(cfg sandbox.Config) sandbox.Executor {
		return stubSandboxExecutor{result: map[string]interface{}{"msg": "sandbox-smoke-test"}}
	}
	t.Cleanup(func() { newSubprocessSandboxExecutor = original })

	cfg := config.DefaultConfig()
	cfg.P2P.ToolIsolation.Enabled = true
	cfg.P2P.ToolIsolation.Container.Enabled = false

	cmd := newSandboxTestCmd(testutil.FakeBootLoader(t, cfg))
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Using subprocess sandbox")
	assert.Contains(t, out, "Smoke test passed: map[msg:sandbox-smoke-test]")
}

func TestSandboxCleanupCmd_WritesToCommandWriter(t *testing.T) {
	original := newSandboxDockerRuntime
	newSandboxDockerRuntime = func() (sandboxCleanupRuntime, error) {
		return stubCleanupRuntime{available: true}, nil
	}
	t.Cleanup(func() { newSandboxDockerRuntime = original })

	cmd := newSandboxCleanupCmd(func() (*bootstrap.Result, error) {
		return &bootstrap.Result{}, nil
	})
	out, err := executeP2PCmd(t, cmd)
	require.NoError(t, err)
	assert.Contains(t, out, "Orphaned sandbox containers cleaned up.")
}
