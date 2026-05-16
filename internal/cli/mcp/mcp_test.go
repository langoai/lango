package mcp

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
	mcplib "github.com/langoai/lango/internal/mcp"
	"github.com/langoai/lango/internal/testutil"
)

func executeMCPCmd(t *testing.T, cmd *cobra.Command, args ...string) (string, error) {
	t.Helper()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	err := cmd.Execute()
	return out.String(), err
}

func TestMCPGet_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"filesystem": {
			Enabled: boolPtr(false),
			Command: "npx",
			Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
		},
	}
	cmd := NewMCPCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	out, err := executeMCPCmd(t, cmd, "get", "filesystem")

	require.NoError(t, err)
	assert.Contains(t, out, "Server: filesystem")
	assert.Contains(t, out, "Transport:    stdio")
	assert.Contains(t, out, "(server is disabled)")
}

func TestMCPList_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"filesystem": {
			Command: "npx",
			Args:    []string{"@modelcontextprotocol/server-filesystem", "/tmp"},
		},
	}
	cmd := NewMCPCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	out, err := executeMCPCmd(t, cmd, "list")

	require.NoError(t, err)
	assert.Contains(t, out, "NAME")
	assert.Contains(t, out, "filesystem")
	assert.Contains(t, out, "stdio")
}

func TestMCPList_EmptyStateWritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MCP.Servers = map[string]config.MCPServerConfig{}
	cmd := NewMCPCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	out, err := executeMCPCmd(t, cmd, "list")

	require.NoError(t, err)
	assert.Contains(t, out, "No MCP servers configured.")
	assert.Contains(t, out, "lango mcp add")
}

func TestMCPEnableDisable_WriteToCommandOutput(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	projectFile := filepath.Join(tmp, ".lango-mcp.json")
	require.NoError(t, mcplib.SaveMCPFile(projectFile, map[string]config.MCPServerConfig{
		"filesystem": {
			Command: "npx",
			Enabled: boolPtr(false),
		},
	}))

	cfg := config.DefaultConfig()
	cmd := NewMCPCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	enableOut, err := executeMCPCmd(t, cmd, "enable", "filesystem", "--scope", "project")
	require.NoError(t, err)
	assert.Contains(t, enableOut, `MCP server "filesystem" enabled.`)

	disableOut, err := executeMCPCmd(t, cmd, "disable", "filesystem", "--scope", "project")
	require.NoError(t, err)
	assert.Contains(t, disableOut, `MCP server "filesystem" disabled.`)
}

func TestMCPAddRemove_WriteToCommandOutput(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })

	cfg := config.DefaultConfig()
	cmd := NewMCPCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	addOut, err := executeMCPCmd(t, cmd,
		"add", "filesystem",
		"--scope", "project",
		"--type", "stdio",
		"--command", "npx",
		"--args", "@modelcontextprotocol/server-filesystem,/tmp",
	)
	require.NoError(t, err)
	assert.Contains(t, addOut, `MCP server "filesystem" added to project scope`)
	assert.Contains(t, addOut, "Transport: stdio")

	removeOut, err := executeMCPCmd(t, cmd, "remove", "filesystem", "--scope", "project")
	require.NoError(t, err)
	assert.Contains(t, removeOut, `MCP server "filesystem" removed from project scope.`)
}

func TestMCPTest_WritesToCommandOutput(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.MCP.Enabled = true
	cfg.MCP.Servers = map[string]config.MCPServerConfig{
		"remote-api": {
			Transport: "http",
			URL:       "http://127.0.0.1:1/mcp",
		},
	}
	cmd := NewMCPCmd(testutil.FakeCfgLoader(cfg), testutil.FailBootLoader(assert.AnError))

	out, err := executeMCPCmd(t, cmd, "test", "remote-api")

	require.NoError(t, err)
	assert.Contains(t, out, `Testing "remote-api"...`)
	assert.Contains(t, out, "Transport:  http")
	assert.Contains(t, out, "Handshake:  FAILED")
}
