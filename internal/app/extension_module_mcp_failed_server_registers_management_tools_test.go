package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/appinit"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/eventbus"
	"github.com/langoai/lango/internal/mcp"
)

func TestExtensionModuleMCPFailedServerRegistersManagementTools(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))

	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"broken": {
			Transport: "invalid-transport",
		},
	}
	cfg.Observability.Enabled = false

	result, err := (&extensionModule{cfg: cfg, bus: eventbus.New()}).
		Init(context.Background(), staticResolver{})
	require.NoError(t, err)

	mcpc, ok := result.Values[appinit.ProvidesMCP].(*mcpComponents)
	require.True(t, ok)
	require.NotNil(t, mcpc)
	require.Equal(t, 1, mcpc.manager.ServerCount())

	conn, ok := mcpc.manager.GetConnection("broken")
	require.True(t, ok)
	assert.Equal(t, mcp.StateFailed, conn.State())

	entries := extensionModuleMCPFailedServerRegistersManagementToolsCatalogEntries(
		result.CatalogEntries,
		"mcp",
	)
	require.Len(t, entries, 2)
	assert.Equal(t, "MCP plugin tools (external servers)", entries[0].Description)
	assert.True(t, entries[0].Enabled)
	assert.Empty(t, entries[0].Tools)

	management := findTool(result.Tools, "mcp_status")
	require.NotNil(t, management)
	status, err := management.Handler(context.Background(), nil)
	require.NoError(t, err)
	assert.Equal(t, "broken: failed", status)

	assert.NotNil(t, findTool(result.Tools, "mcp_tools"))
	assert.Equal(t, "MCP management tools", entries[1].Description)
	assert.True(t, entries[1].Enabled)
	assert.ElementsMatch(t, []string{"mcp_status", "mcp_tools"}, catalogEntryToolNames(entries[1]))
	require.Len(t, result.Components, 1)
	assert.Equal(t, "mcp-manager", result.Components[0].Component.Name())
}

func extensionModuleMCPFailedServerRegistersManagementToolsCatalogEntries(
	entries []appinit.CatalogEntry,
	category string,
) []appinit.CatalogEntry {
	var out []appinit.CatalogEntry
	for _, entry := range entries {
		if entry.Category == category {
			out = append(out, entry)
		}
	}
	return out
}
