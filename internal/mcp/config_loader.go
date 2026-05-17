package mcp

import (
	"encoding/json"
	"errors"
	"fmt"
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
	merged, err := MergedServersStrict(cfg)
	if err != nil {
		logging.App().Warnw("failed to load scoped MCP config", "error", err)
	}
	return merged
}

// MergedServersStrict loads and merges MCP server configs from all scopes.
// Missing scoped files are optional; present files that cannot load fail closed.
func MergedServersStrict(cfg *config.MCPConfig) (map[string]config.MCPServerConfig, error) {
	merged := make(map[string]config.MCPServerConfig)

	// 1. Profile-level servers (from config DB)
	if cfg != nil {
		for name, srv := range cfg.Servers {
			merged[name] = srv
		}
	}

	// 2. User-level (~/.lango/mcp.json)
	if home, err := os.UserHomeDir(); err == nil {
		userPath := filepath.Join(home, ".lango", "mcp.json")
		servers, err := loadScopedMCPFile("user", userPath)
		if err != nil {
			return merged, err
		}
		if servers != nil {
			for name, srv := range servers {
				merged[name] = srv
			}
		}
	}

	// 3. Project-level (.lango-mcp.json)
	projectPath := ".lango-mcp.json"
	servers, err := loadScopedMCPFile("project", projectPath)
	if err != nil {
		return merged, err
	}
	if servers != nil {
		for name, srv := range servers {
			merged[name] = srv
		}
	}

	return merged, nil
}

// LoadScopedMCPFile reads a scoped MCP config file. Missing files return nil.
func LoadScopedMCPFile(scope, path string) (map[string]config.MCPServerConfig, error) {
	return loadScopedMCPFile(scope, path)
}

func loadScopedMCPFile(scope, path string) (map[string]config.MCPServerConfig, error) {
	servers, err := loadMCPFile(path)
	if err == nil {
		return servers, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return nil, fmt.Errorf("load %s MCP config %s: %w", scope, path, err)
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
	if err := validateMCPFileServers(fc.MCPServers); err != nil {
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

func validateMCPFileServers(servers map[string]config.MCPServerConfig) error {
	for name, srv := range servers {
		if !config.ValidMCPTransports[srv.Transport] {
			return fmt.Errorf("mcpServers.%s.transport %q is not supported (must be stdio, http, or sse)", name, srv.Transport)
		}
		switch srv.Transport {
		case "", "stdio":
			if srv.Command == "" {
				return fmt.Errorf("mcpServers.%s.command is required for stdio transport", name)
			}
		case "http", "sse":
			if srv.URL == "" {
				return fmt.Errorf("mcpServers.%s.url is required for %s transport", name, srv.Transport)
			}
		}
		switch srv.SafetyLevel {
		case "", "safe", "moderate", "dangerous":
		default:
			return fmt.Errorf("mcpServers.%s.safetyLevel %q is not supported (must be safe, moderate, or dangerous)", name, srv.SafetyLevel)
		}
	}
	return nil
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
