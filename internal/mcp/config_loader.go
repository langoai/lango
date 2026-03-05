package mcp

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/logging"
)

// mcpFileConfig is the JSON schema for .lango-mcp.json / ~/.lango/mcp.json files.
type mcpFileConfig struct {
	MCPServers map[string]config.MCPServerConfig `json:"mcpServers"`
}

// MergedServers loads and merges MCP server configs from multiple scopes:
//  1. Profile config (cfg.Servers, lowest priority)
//  2. User-level config (~/.lango/mcp.json)
//  3. Project-level config (.lango-mcp.json, highest priority)
//
// Later scopes override earlier ones on a per-server-name basis.
func MergedServers(cfg *config.MCPConfig) map[string]config.MCPServerConfig {
	merged := make(map[string]config.MCPServerConfig)

	// 1. Profile-level servers (from config DB)
	for name, srv := range cfg.Servers {
		merged[name] = srv
	}

	// 2. User-level (~/.lango/mcp.json)
	if home, err := os.UserHomeDir(); err == nil {
		userPath := filepath.Join(home, ".lango", "mcp.json")
		if servers, err := loadMCPFile(userPath); err == nil {
			for name, srv := range servers {
				merged[name] = srv
			}
		}
	}

	// 3. Project-level (.lango-mcp.json)
	projectPath := ".lango-mcp.json"
	if servers, err := loadMCPFile(projectPath); err == nil {
		for name, srv := range servers {
			merged[name] = srv
		}
	}

	return merged
}

// loadMCPFile reads an MCP config file and returns the server map.
func loadMCPFile(path string) (map[string]config.MCPServerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fc mcpFileConfig
	if err := json.Unmarshal(data, &fc); err != nil {
		logging.App().Warnw("invalid MCP config file", "path", path, "error", err)
		return nil, err
	}

	// Apply env expansion to loaded configs
	for name, srv := range fc.MCPServers {
		srv.Env = ExpandEnvMap(srv.Env)
		for k, v := range srv.Headers {
			srv.Headers[k] = ExpandEnv(v)
		}
		fc.MCPServers[name] = srv
	}

	logging.App().Infow("loaded MCP config file", "path", path, "servers", len(fc.MCPServers))
	return fc.MCPServers, nil
}

// SaveMCPFile writes MCP server configs to a JSON file.
func SaveMCPFile(path string, servers map[string]config.MCPServerConfig) error {
	fc := mcpFileConfig{MCPServers: servers}
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// LoadMCPFile reads an MCP config file and returns the server map (exported).
func LoadMCPFile(path string) (map[string]config.MCPServerConfig, error) {
	return loadMCPFile(path)
}
