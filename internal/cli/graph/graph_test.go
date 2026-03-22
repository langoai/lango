package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/testutil"
)

func TestNewGraphCmd_Structure(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	require.NotNil(t, cmd)
	assert.Equal(t, "graph", cmd.Use)
	assert.NotEmpty(t, cmd.Short)
}

func TestNewGraphCmd_Subcommands(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	expected := []string{"status", "query", "stats", "clear", "add", "export", "import"}
	subCmds := make(map[string]bool, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		subCmds[sub.Name()] = true
	}

	for _, name := range expected {
		assert.True(t, subCmds[name], "missing subcommand: %s", name)
	}
}

func TestStatusCmd_GraphDisabled(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status")
	assert.Contains(t, result.Stdout, "Knowledge Graph Status")
	assert.Contains(t, result.Stdout, "false")
}

func TestStatusCmd_GraphDisabledJSON(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmdOK(t, cmd, "status", "--json")
	assert.Contains(t, result.Stdout, `"enabled": false`)
	assert.Contains(t, result.Stdout, `"triple_count": 0`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cmd := NewGraphCmd(testutil.FailCfgLoader(assert.AnError))
	result := testutil.ExecCmd(t, cmd, "status")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "load config")
}

func TestQueryCmd_MissingFlags(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	// Neither --subject nor --object provided
	result := testutil.ExecCmd(t, cmd, "query")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "at least one of --subject or --object is required")
}

func TestQueryCmd_PredicateWithoutSubject(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	// Provide --object so the first check passes, but --predicate without --subject should fail.
	result := testutil.ExecCmd(t, cmd, "query", "--predicate", "knows", "--object", "Bob")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "--predicate requires --subject")
}

func TestAddCmd_MissingRequiredFlags(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	// Missing all required flags
	result := testutil.ExecCmd(t, cmd, "add")
	require.Error(t, result.Err)
}

func TestExportCmd_InvalidFormat(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "export", "--format", "xml")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "must be json or csv")
}

func TestImportCmd_MissingFileArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "import")
	require.Error(t, result.Err)
}

func TestImportCmd_NonexistentFile(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "import", "/nonexistent/file.json")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "read file")
}

func TestStatsCmd_GraphDisabledError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "stats")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "graph store is not enabled")
}

func TestStatusCmd_GraphEnabledNoPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = ""
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	result := testutil.ExecCmd(t, cmd, "status")
	require.Error(t, result.Err)
	assert.Contains(t, result.Err.Error(), "graph database path is not configured")
}
