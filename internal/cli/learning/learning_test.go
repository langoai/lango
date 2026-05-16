package learning

import (
	"bytes"
	"testing"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func executeLearningCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

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

	out, err := executeLearningCmd(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Learning Status")
	assert.Contains(t, out, "Knowledge Enabled")
}

func TestStatusCmd_JSONOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Knowledge.Enabled = true
	cfg.Embedding.Provider = "local"
	cfg.Embedding.Model = "nomic"
	cfg.Embedding.RAG.Enabled = true
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLearningCmd(t, cmd, "status", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"knowledge_enabled": true`)
	assert.Contains(t, out, `"embedding_provider": "local"`)
	assert.Contains(t, out, `"rag_enabled": true`)
}

func TestStatusCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(func() (*config.Config, error) {
		t.Fatal("config loader should not be called for invalid output")
		return nil, nil
	}, testutil.FakeBootLoader(t, cfg))

	_, err := executeLearningCmd(t, cmd, "status", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FailCfgLoader(assert.AnError), testutil.FakeBootLoader(t, cfg))

	_, err := executeLearningCmd(t, cmd, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestHistoryCmd_EmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLearningCmd(t, cmd, "history")
	require.NoError(t, err)
	assert.Contains(t, out, "No learning entries found.")
}

func TestHistoryCmd_JSONEmptyDB(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FakeBootLoader(t, cfg))

	out, err := executeLearningCmd(t, cmd, "history", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, "[]")
}

func TestHistoryCmd_InvalidOutputFailsBeforeBootLoad(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), func() (*bootstrap.Result, error) {
		t.Fatal("boot loader should not be called for invalid output")
		return nil, nil
	})

	_, err := executeLearningCmd(t, cmd, "history", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestHistoryCmd_BootError(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewLearningCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	_, err := executeLearningCmd(t, cmd, "history")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bootstrap")
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
