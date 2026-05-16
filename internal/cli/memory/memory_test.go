package memory

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/langoai/lango/internal/agentmemory"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
	"github.com/langoai/lango/internal/testutil"
)

func executeMemoryCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func executeMemoryCmdWithInput(t *testing.T, cmd *cobra.Command, input string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(bytes.NewBufferString(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func tempMemoryConfig(t *testing.T) *config.Config {
	t.Helper()
	cfg := config.DefaultConfig()
	cfg.Session.DatabasePath = filepath.Join(t.TempDir(), "memory-test.db")
	return cfg
}

func seedAgentMemory(t *testing.T, cfg *config.Config, entries ...*agentmemory.Entry) {
	t.Helper()
	store, err := session.NewEntStore(cfg.Session.DatabasePath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	memStore := agentmemory.NewEntStore(store.Client())
	for _, entry := range entries {
		require.NoError(t, memStore.Save(entry))
	}
}

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
	cfg := tempMemoryConfig(t)
	cfg.AgentMemory.Enabled = true
	seedAgentMemory(t, cfg,
		&agentmemory.Entry{AgentName: "librarian", Key: "k1", Content: "remember A", Scope: agentmemory.ScopeGlobal, Kind: agentmemory.KindFact, Confidence: 0.9},
		&agentmemory.Entry{AgentName: "operator", Key: "k2", Content: "remember B", Scope: agentmemory.ScopeGlobal, Kind: agentmemory.KindFact, Confidence: 0.8},
	)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "agents")
	require.NoError(t, err)
	assert.Contains(t, out, "AGENT")
	assert.Contains(t, out, "librarian")
	assert.Contains(t, out, "operator")
}

func TestAgentsCmd_JSONOutput(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cfg.AgentMemory.Enabled = true
	seedAgentMemory(t, cfg,
		&agentmemory.Entry{AgentName: "planner", Key: "k1", Content: "remember C", Scope: agentmemory.ScopeGlobal, Kind: agentmemory.KindFact, Confidence: 0.9},
	)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "agents", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"agent_name": "planner"`)
	assert.Contains(t, out, `"entry_count": 1`)
}

func TestAgentsCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewMemoryCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeMemoryCmd(t, cmd, "agents", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestAgentsCmd_ConfigError(t *testing.T) {
	cmd := NewMemoryCmd(testutil.FailCfgLoader(assert.AnError))
	_, err := executeMemoryCmd(t, cmd, "agents")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestAgentCmd_Enabled(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cfg.AgentMemory.Enabled = true
	seedAgentMemory(t, cfg,
		&agentmemory.Entry{AgentName: "researcher", Key: "k1", Content: "User prefers Go", Scope: agentmemory.ScopeGlobal, Kind: agentmemory.KindPreference, Confidence: 0.9},
		&agentmemory.Entry{AgentName: "researcher", Key: "k2", Content: "Uses BoltDB", Scope: agentmemory.ScopeGlobal, Kind: agentmemory.KindFact, Confidence: 0.8},
	)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "agent", "researcher")
	require.NoError(t, err)
	assert.Contains(t, out, "ID")
	assert.Contains(t, out, "User prefers Go")
	assert.Contains(t, out, "Uses BoltDB")
}

func TestAgentCmd_Disabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = false
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeMemoryCmd(t, cmd, "agent", "researcher")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "agent memory is not enabled")
}

func TestAgentCmd_JSONOutput(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cfg.AgentMemory.Enabled = true
	seedAgentMemory(t, cfg,
		&agentmemory.Entry{AgentName: "planner", Key: "k1", Content: "Plan carefully", Scope: agentmemory.ScopeGlobal, Kind: agentmemory.KindSkill, Confidence: 0.9},
	)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "agent", "planner", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"agent_name": "planner"`)
	assert.Contains(t, out, `"content": "Plan carefully"`)
}

func TestAgentCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewMemoryCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeMemoryCmd(t, cmd, "agent", "planner", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestAgentCmd_LimitFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	for _, sub := range cmd.Commands() {
		if sub.Name() == "agent" {
			f := sub.Flags().Lookup("limit")
			require.NotNil(t, f, "agent command should have --limit flag")
			assert.Equal(t, "20", f.DefValue)
			return
		}
	}
	t.Fatal("agent subcommand not found")
}

func TestAgentCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.AgentMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeMemoryCmd(t, cmd, "agent")
	require.Error(t, err)
}

func TestListCmd_MissingSessionFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeMemoryCmd(t, cmd, "list")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session")
}

func TestStatusCmd_MissingSessionFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeMemoryCmd(t, cmd, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "session")
}

func TestListCmd_WritesEmptyStateToCommandWriter(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "list", "--session", "missing-session")
	require.NoError(t, err)
	assert.Contains(t, out, "No entries found.")
}

func TestListCmd_JSONWritesToCommandWriter(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "list", "--session", "missing-session", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "[]")
}

func TestListCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewMemoryCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeMemoryCmd(t, cmd, "list", "--session", "missing-session", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestStatusCmd_WritesToCommandWriter(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cfg.ObservationalMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "status", "--session", "missing-session")
	require.NoError(t, err)
	assert.Contains(t, out, "Observational Memory Status (session: missing-session)")
	assert.Contains(t, out, "Observations:")
}

func TestStatusCmd_JSONWritesToCommandWriter(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cfg.ObservationalMemory.Enabled = true
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "status", "--session", "missing-session", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"observations": 0`)
	assert.Contains(t, out, `"enabled": true`)
}

func TestStatusCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewMemoryCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeMemoryCmd(t, cmd, "status", "--session", "missing-session", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestClearCmd_MissingArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeMemoryCmd(t, cmd, "clear")
	require.Error(t, err)
}

func TestClearCmd_AbortUsesCommandStreams(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmdWithInput(t, cmd, "n\n", "clear", "missing-session")

	require.NoError(t, err)
	assert.Contains(t, out, "This will delete all observations and reflections for session 'missing-session'.")
	assert.Contains(t, out, "Continue? [y/N]")
	assert.Contains(t, out, "Aborted.")
}

func TestClearCmd_ConfirmUsesCommandStreams(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmdWithInput(t, cmd, "y\n", "clear", "missing-session")

	require.NoError(t, err)
	assert.Contains(t, out, "Continue? [y/N]")
	assert.Contains(t, out, "Cleared all memory entries for session 'missing-session'.")
}

func TestClearCmd_ForceWritesToCommandOutput(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmd(t, cmd, "clear", "missing-session", "--force")

	require.NoError(t, err)
	assert.Contains(t, out, "Cleared all memory entries for session 'missing-session'.")
}

func TestClearCmd_EOFAbortsWithoutDeleting(t *testing.T) {
	cfg := tempMemoryConfig(t)
	cmd := NewMemoryCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeMemoryCmdWithInput(t, cmd, "", "clear", "missing-session")

	require.NoError(t, err)
	assert.Contains(t, out, "This will delete all observations and reflections for session 'missing-session'.")
	assert.Contains(t, out, "Continue? [y/N]: ")
	assert.Contains(t, out, "Aborted.")
}
