package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	graphstore "github.com/langoai/lango/internal/graph"
	"github.com/langoai/lango/internal/testutil"
)

func executeGraphCommand(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func executeGraphCommandWithInput(t *testing.T, cmd *cobra.Command, input string, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetIn(bytes.NewBufferString(input))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

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

	out, err := executeGraphCommand(t, cmd, "status")
	require.NoError(t, err)
	assert.Contains(t, out, "Knowledge Graph Status")
	assert.Contains(t, out, "false")
}

func TestStatusCmd_GraphDisabledJSON(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "status", "--output", "json")
	require.NoError(t, err)
	assert.Contains(t, out, `"enabled": false`)
	assert.Contains(t, out, `"triple_count": 0`)
}

func TestStatusCmd_TableWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "status")

	require.NoError(t, err)
	assert.Contains(t, out, "Knowledge Graph Status")
	assert.Contains(t, out, "Enabled:  false")
}

func TestStatusCmd_JSONWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "status", "--output", "json")

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, false, payload["enabled"])
	assert.EqualValues(t, 0, payload["triple_count"])
}

func TestStatusCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewGraphCmd(testutil.FailCfgLoader(assert.AnError))
	_, err := executeGraphCommand(t, cmd, "status", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestStatusCmd_ConfigError(t *testing.T) {
	cmd := NewGraphCmd(testutil.FailCfgLoader(assert.AnError))
	_, err := executeGraphCommand(t, cmd, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "load config")
}

func TestQueryCmd_MissingFlags(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	// Neither --subject nor --object provided
	_, err := executeGraphCommand(t, cmd, "query")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one of --subject or --object is required")
}

func TestQueryCmd_PredicateWithoutSubject(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	// Provide --object so the first check passes, but --predicate without --subject should fail.
	_, err := executeGraphCommand(t, cmd, "query", "--predicate", "knows", "--object", "Bob")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "--predicate requires --subject")
}

func TestQueryCmd_WritesEmptyStateToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "query", "--subject", "missing")

	require.NoError(t, err)
	assert.Contains(t, out, "No triples found.")
}

func TestQueryCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "query", "--subject", "Go")

	require.NoError(t, err)
	assert.Contains(t, out, "SUBJECT")
	assert.Contains(t, out, "Go")
	assert.Contains(t, out, "is_a")
}

func TestQueryCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "query", "--subject", "Go", "--output", "json")

	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	require.Len(t, payload, 1)
	assert.Equal(t, "Go", payload[0]["Subject"])
	assert.Equal(t, "is_a", payload[0]["Predicate"])
	assert.Equal(t, "language", payload[0]["Object"])
}

func TestQueryCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewGraphCmd(testutil.FailCfgLoader(assert.AnError))
	_, err := executeGraphCommand(t, cmd, "query", "--subject", "Go", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestAddCmd_MissingRequiredFlags(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	// Missing all required flags
	_, err := executeGraphCommand(t, cmd, "add")
	require.Error(t, err)
}

func TestAddCmd_WritesTextOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd,
		"add",
		"--subject", "Go",
		"--predicate", "is_a",
		"--object", "language",
	)

	require.NoError(t, err)
	assert.Contains(t, out, "Added triple: (Go) -[is_a]-> (language)")
}

func TestAddCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd,
		"add",
		"--subject", "Go",
		"--predicate", "is_a",
		"--object", "language",
		"--output", "json",
	)

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.Equal(t, "Go", payload["Subject"])
	assert.Equal(t, "is_a", payload["Predicate"])
	assert.Equal(t, "language", payload["Object"])
}

func TestAddCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewGraphCmd(testutil.FailCfgLoader(assert.AnError))
	_, err := executeGraphCommand(t, cmd,
		"add",
		"--subject", "Go",
		"--predicate", "is_a",
		"--object", "language",
		"--output", "yaml",
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestExportCmd_InvalidFormat(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeGraphCommand(t, cmd, "export", "--format", "xml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be json or csv")
}

func TestClearCmd_AbortUsesCommandStreams(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommandWithInput(t, cmd, "n\n", "clear")

	require.NoError(t, err)
	assert.Contains(t, out, "This will delete all triples from the knowledge graph.")
	assert.Contains(t, out, "Continue? [y/N]")
	assert.Contains(t, out, "Aborted.")
}

func TestClearCmd_ConfirmUsesCommandStreams(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommandWithInput(t, cmd, "y\n", "clear")

	require.NoError(t, err)
	assert.Contains(t, out, "Continue? [y/N]")
	assert.Contains(t, out, "Cleared all triples from the knowledge graph.")
}

func TestClearCmd_ForceWritesToCommandOutput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "clear", "--force")

	require.NoError(t, err)
	assert.Contains(t, out, "Cleared all triples from the knowledge graph.")
}

func TestClearCmd_EOFAbortsWithoutClearing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommandWithInput(t, cmd, "", "clear")

	require.NoError(t, err)
	assert.Contains(t, out, "This will delete all triples from the knowledge graph.")
	assert.Contains(t, out, "Continue? [y/N]: ")
	assert.Contains(t, out, "Aborted.")
}

func TestExportCmd_JSONWritesToCommandOutput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "export", "--format", "json")

	require.NoError(t, err)
	var payload []map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	require.Len(t, payload, 1)
	assert.Equal(t, "Go", payload[0]["Subject"])
}

func TestExportCmd_CSVWritesToCommandOutput(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: "is_a",
		Object:    "language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "export", "--format", "csv")

	require.NoError(t, err)
	assert.Contains(t, out, "subject,predicate,object")
	assert.Contains(t, out, "Go,is_a,language")
}

func TestImportCmd_MissingFileArg(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeGraphCommand(t, cmd, "import")
	require.Error(t, err)
}

func TestImportCmd_NonexistentFile(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeGraphCommand(t, cmd, "import", "/nonexistent/file.json")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "read file")
}

func TestImportCmd_WritesEmptyStateToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	importFile := filepath.Join(t.TempDir(), "empty.json")
	require.NoError(t, os.WriteFile(importFile, []byte("[]"), 0o600))

	out, err := executeGraphCommand(t, cmd, "import", importFile)

	require.NoError(t, err)
	assert.Contains(t, out, "No triples to import.")
}

func TestImportCmd_JSONWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = filepath.Join(t.TempDir(), "graph.db")
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	importFile := filepath.Join(t.TempDir(), "triples.json")
	require.NoError(t, os.WriteFile(importFile, []byte(`[{"Subject":"Alice","Predicate":"knows","Object":"Bob"}]`), 0o600))

	out, err := executeGraphCommand(t, cmd, "import", importFile, "--output", "json")

	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.EqualValues(t, 1, payload["imported"])
}

func TestImportCmd_InvalidOutputFailsBeforeFileRead(t *testing.T) {
	cfg := config.DefaultConfig()
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeGraphCommand(t, cmd, "import", "/nonexistent/file.json", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}

func TestStatsCmd_GraphDisabledError(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = false
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeGraphCommand(t, cmd, "stats")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "graph store is not enabled")
}

func TestStatusCmd_GraphEnabledNoPath(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = ""
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	_, err := executeGraphCommand(t, cmd, "status")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "graph database path is not configured")
}

func TestStatsCmd_WritesTableOutputToCommandWriter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: graphstore.RelatedTo,
		Object:    "Language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "stats")
	require.NoError(t, err)
	assert.Contains(t, out, "Knowledge Graph Statistics")
	assert.Contains(t, out, "related_to")
}

func TestStatsCmd_WritesJSONOutputToCommandWriter(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "graph.db")
	store, err := graphstore.NewBoltStore(dbPath)
	require.NoError(t, err)
	require.NoError(t, store.AddTriple(context.Background(), graphstore.Triple{
		Subject:   "Go",
		Predicate: graphstore.RelatedTo,
		Object:    "Language",
	}))
	require.NoError(t, store.Close())

	cfg := config.DefaultConfig()
	cfg.Graph.Enabled = true
	cfg.Graph.DatabasePath = dbPath
	cmd := NewGraphCmd(testutil.FakeCfgLoader(cfg))

	out, err := executeGraphCommand(t, cmd, "stats", "--output", "json")
	require.NoError(t, err)
	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(out), &payload))
	assert.EqualValues(t, 1, payload["total_triples"])
}

func TestStatsCmd_InvalidOutputFailsBeforeConfigLoad(t *testing.T) {
	cmd := NewGraphCmd(testutil.FailCfgLoader(assert.AnError))
	_, err := executeGraphCommand(t, cmd, "stats", "--output", "yaml")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `unknown output format "yaml"`)
}
