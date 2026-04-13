package memory

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func TestNewMemoryCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "memory", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewMemoryCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	expected := []string{"list", "status", "clear", "agents", "agent"}
	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestAgentsCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "agents")
	assert.Contains(t, result.Stdout, "Agent Memory")
	assert.Contains(t, result.Stdout, "Enabled: true")
}

func TestAgentsCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = false
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "agents", "--json")
	assert.Contains(t, result.Stdout, `"enabled": false`)
	assert.Contains(t, result.Stdout, `"note"`)
}

func TestAgentsCmd_ConfigError(t *testing.T) {
	cmd := NewMemoryCmd(testutil.FailCfgLoader(assert.AnError))
	result := testutil.ExecCmd(t, cmd, "agents")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "load config")
}

func TestAgentCmd_Enabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "agent", "researcher")
	assert.Contains(t, result.Stdout, "Agent Memory: researcher")
}

func TestAgentCmd_Disabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = false
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "agent", "researcher")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "agent memory is not enabled")
}

func TestAgentCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "agent", "planner", "--json")
	assert.Contains(t, result.Stdout, `"agent_name": "planner"`)
}

func TestAgentCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "agent")
	require.Error(t, result.Err)
}

func TestListCmd_MissingSessionFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "list")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "session")
}

func TestStatusCmd_MissingSessionFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "status")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "session")
}

func TestClearCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "clear")
	require.Error(t, result.Err)
}
