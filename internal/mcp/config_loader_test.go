package mcp

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/langoai/lango/internal/config"
)

func TestLoadMCPFile_ValidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")

	content := `{
		"mcpServers": {
			"server-a": {
				"transport": "stdio",
				"command": "npx",
				"args": ["-y", "@modelcontextprotocol/server-a"]
			},
			"server-b": {
				"transport": "http",
				"url": "http://localhost:3000/mcp"
			}
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	servers, err := LoadMCPFile(path)
	require.NoError(t, err)
	require.Len(t, servers, 2)

	assert.Equal(t, "stdio", servers["server-a"].Transport)
	assert.Equal(t, "npx", servers["server-a"].Command)
	assert.Equal(t, []string{"-y", "@modelcontextprotocol/server-a"}, servers["server-a"].Args)

	assert.Equal(t, "http", servers["server-b"].Transport)
	assert.Equal(t, "http://localhost:3000/mcp", servers["server-b"].URL)
}

func TestLoadMCPFile_NotFound(t *testing.T) {
	t.Parallel()

	_, err := LoadMCPFile("/nonexistent/path/mcp.json")
	assert.Error(t, err)
}

func TestLoadMCPFile_InvalidJSON(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")
	require.NoError(t, os.WriteFile(path, []byte(`{bad json`), 0644))

	_, err := LoadMCPFile(path)
	assert.Error(t, err)
}

func TestLoadMCPFile_EnvExpansion(t *testing.T) {
	os.Setenv("MCP_TEST_TOKEN", "secret-token")
	defer os.Unsetenv("MCP_TEST_TOKEN")

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")

	content := `{
		"mcpServers": {
			"test-srv": {
				"transport": "http",
				"url": "http://localhost:3000",
				"headers": {
					"Authorization": "Bearer ${MCP_TEST_TOKEN}"
				},
				"env": {
					"TOKEN": "${MCP_TEST_TOKEN}"
				}
			}
		}
	}`
	require.NoError(t, os.WriteFile(path, []byte(content), 0644))

	servers, err := LoadMCPFile(path)
	require.NoError(t, err)
	require.Contains(t, servers, "test-srv")

	assert.Equal(t, "Bearer secret-token", servers["test-srv"].Headers["Authorization"])
	assert.Equal(t, "secret-token", servers["test-srv"].Env["TOKEN"])
}

func TestLoadMCPFile_EmptyServers(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"mcpServers": {}}`), 0644))

	servers, err := LoadMCPFile(path)
	require.NoError(t, err)
	assert.Empty(t, servers)
}

func TestSaveMCPFile(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "mcp.json")

	servers := map[string]config.MCPServerConfig{
		"my-server": {
			Transport: "stdio",
			Command:   "node",
			Args:      []string{"server.js"},
		},
	}

	require.NoError(t, SaveMCPFile(path, servers))

	// Verify round-trip
	loaded, err := LoadMCPFile(path)
	require.NoError(t, err)
	require.Contains(t, loaded, "my-server")
	assert.Equal(t, "stdio", loaded["my-server"].Transport)
	assert.Equal(t, "node", loaded["my-server"].Command)
	assert.Equal(t, []string{"server.js"}, loaded["my-server"].Args)
}

func TestMergedServers_ProfilePriority(t *testing.T) {
	t.Parallel()

	cfg := &config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"profile-srv": {
				Transport: "stdio",
				Command:   "profile-cmd",
			},
		},
	}

	merged := MergedServers(cfg)
	assert.Contains(t, merged, "profile-srv")
	assert.Equal(t, "profile-cmd", merged["profile-srv"].Command)
}

func TestMergedServersStrict_IgnoresMissingScopedFiles(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))

	cfg := &config.MCPConfig{
		Servers: map[string]config.MCPServerConfig{
			"profile-srv": {
				Transport: "stdio",
				Command:   "profile-cmd",
			},
		},
	}

	merged, err := MergedServersStrict(cfg)

	require.NoError(t, err)
	require.Contains(t, merged, "profile-srv")
	assert.Equal(t, "profile-cmd", merged["profile-srv"].Command)
}

func TestMergedServersStrict_InvalidUserConfigReturnsScopeAndPath(t *testing.T) {
	tmp := t.TempDir()
	home := filepath.Join(tmp, "home")
	userDir := filepath.Join(home, ".lango")
	require.NoError(t, os.MkdirAll(userDir, 0700))
	userPath := filepath.Join(userDir, "mcp.json")
	require.NoError(t, os.WriteFile(userPath, []byte(`{bad json`), 0644))
	t.Setenv("HOME", home)

	_, err := MergedServersStrict(&config.MCPConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user")
	assert.Contains(t, err.Error(), userPath)
}

func TestMergedServersStrict_InvalidProjectConfigReturnsScopeAndPath(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))
	projectPath := filepath.Join(tmp, ".lango-mcp.json")
	require.NoError(t, os.WriteFile(projectPath, []byte(`{bad json`), 0644))

	_, err = MergedServersStrict(&config.MCPConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
	assert.Contains(t, err.Error(), ".lango-mcp.json")
}

func TestMergedServersStrict_InvalidProjectServerConfigReturnsScopeAndPath(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".lango-mcp.json"), []byte(`{
		"mcpServers": {
			"remote-api": {
				"transport": "http"
			}
		}
	}`), 0644))

	_, err = MergedServersStrict(&config.MCPConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
	assert.Contains(t, err.Error(), ".lango-mcp.json")
	assert.Contains(t, err.Error(), "remote-api")
	assert.Contains(t, err.Error(), "url is required")
}

func TestMergedServersStrict_InvalidSafetyLevelReturnsScopeAndPath(t *testing.T) {
	tmp := t.TempDir()
	origWD, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmp))
	t.Cleanup(func() { _ = os.Chdir(origWD) })
	t.Setenv("HOME", filepath.Join(tmp, "home"))
	require.NoError(t, os.WriteFile(filepath.Join(tmp, ".lango-mcp.json"), []byte(`{
		"mcpServers": {
			"filesystem": {
				"transport": "stdio",
				"command": "npx",
				"safetyLevel": "trusted"
			}
		}
	}`), 0644))

	_, err = MergedServersStrict(&config.MCPConfig{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "project")
	assert.Contains(t, err.Error(), ".lango-mcp.json")
	assert.Contains(t, err.Error(), "filesystem")
	assert.Contains(t, err.Error(), "safetyLevel")
}
