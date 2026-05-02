package agent

import (
	"testing"

	"github.com/langoai/lango/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentStatus_DynamicRuntimeDoesNotRequireBuiltinSubAgents(t *testing.T) {
	originalLoader := loadAgentRegistryCounts
	loadAgentRegistryCounts = func(_ *config.Config) (int, int, int) {
		return 0, 0, 0
	}
	t.Cleanup(func() {
		loadAgentRegistryCounts = originalLoader
	})

	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	assert.Contains(t, output, `"teammate_runtime": "dynamic-v1"`)
	assert.Contains(t, output, `"builtin": 0`)
	assert.Contains(t, output, `"active": 0`)
}

func TestStatusCmd_TableHintsBackgroundRequirementForDynamicRuntime(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = false
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	assert.Contains(t, output, "Enable background.enabled to report dynamic-v1 teammate runtime.")
}

func TestStatusCmd_JSONOmitsTeammateRuntimeWithoutAutomation(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Cron.Enabled = false
		cfg.Background.Enabled = false
		cfg.Workflow.Enabled = false
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	assert.NotContains(t, output, `"teammate_runtime"`)
}

func TestStatusCmd_JSONOmitsTeammateRuntimeInSingleAgentMode(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = false
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	assert.NotContains(t, output, `"teammate_runtime"`)
}

func TestStatusCmd_TableIncludesTeammateRuntimeWhenBackgroundEnabled(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Background.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	assert.Contains(t, output, "Teammate Runtime:  dynamic-v1")
}

func TestStatusCmd_JSONOmitsTeammateRuntimeWithoutBackgroundSubmitter(t *testing.T) {
	cmd := newStatusCmd(func() (*config.Config, error) {
		cfg := config.DefaultConfig()
		cfg.Agent.MultiAgent = true
		cfg.Cron.Enabled = true
		cfg.Background.Enabled = false
		cfg.Workflow.Enabled = true
		cfg.Agent.Provider = "anthropic"
		cfg.Agent.Model = "claude-4"
		return cfg, nil
	})
	cmd.SetArgs([]string{"--json"})

	output, err := captureStdout(t, cmd.Execute)
	require.NoError(t, err)

	assert.NotContains(t, output, `"teammate_runtime"`)
}
