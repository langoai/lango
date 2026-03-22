package learning

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func TestNewLearningCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "learning", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewLearningCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	expected := []string{"status", "history"}
	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestStatusCmd_HappyPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true
	cfg.Graph.Enabled = false
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Learning Status")
	assert.Contains(t, result.Stdout, "Knowledge Enabled")
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true
	cfg.Embedding.Provider = "local"
	cfg.Embedding.Model = "nomic"
	cfg.Embedding.RAG.Enabled = true
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "status", "--json")
	assert.Contains(t, result.Stdout, `"knowledge_enabled": true`)
	assert.Contains(t, result.Stdout, `"embedding_provider": "local"`)
	assert.Contains(t, result.Stdout, `"rag_enabled": true`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FailCfgLoader(assert.AnError), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmd(t, cmd, "status")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "load config")
}

func TestHistoryCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "history")
	assert.Contains(t, result.Stdout, "No learning entries found.")
}

func TestHistoryCmd_JSONEmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	result := testutil.ExecCmdOK(t, cmd, "history", "--json")
	assert.Contains(t, result.Stdout, "[]")
}

func TestHistoryCmd_BootError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	result := testutil.ExecCmd(t, cmd, "history")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "bootstrap")
}

func TestHistoryCmd_HasLimitFlag(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	for _, sub := range cmd.Commands() {
		if sub.Name() == "history" {
			f := sub.Flags().Lookup("limit")
			require.NotNil(t, f, "history command should have --limit flag")
			assert.Equal(t, "20", f.DefValue)
			return
		}
	}
	t.Fatal("history subcommand not found")
}
