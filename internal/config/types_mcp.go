package config

import "time"

// MCPConfig defines MCP (Model Context Protocol) server integration settings.
type MCPConfig struct {
	// Enable MCP server integration
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Servers is a map of named MCP server configurations.
	Servers map[string]MCPServerConfig `mapstructure:"servers" json:"servers"`

	// DefaultTimeout for MCP operations (default: 30s)
	DefaultTimeout time.Duration `mapstructure:"defaultTimeout" json:"defaultTimeout"`

	// MaxOutputTokens limits the output size from MCP tool calls (default: 25000)
	MaxOutputTokens int `mapstructure:"maxOutputTokens" json:"maxOutputTokens"`

	// HealthCheckInterval for periodic server health probes (default: 30s)
	HealthCheckInterval time.Duration `mapstructure:"healthCheckInterval" json:"healthCheckInterval"`

	// AutoReconnect enables automatic reconnection on connection loss (default: true)
	AutoReconnect bool `mapstructure:"autoReconnect" json:"autoReconnect"`

	// MaxReconnectAttempts limits reconnection attempts (default: 5)
	MaxReconnectAttempts int `mapstructure:"maxReconnectAttempts" json:"maxReconnectAttempts"`
}

// MCPServerConfig defines a single MCP server connection.
type MCPServerConfig struct {
	// Transport type: "stdio" (default), "http", "sse"
	Transport string `mapstructure:"transport" json:"transport"`

	// Command is the executable for stdio transport.
	Command string `mapstructure:"command" json:"command"`

	// Args are command-line arguments for stdio transport.
	Args []string `mapstructure:"args" json:"args"`

	// Env are environment variables for stdio transport (supports ${VAR} expansion).
	Env map[string]string `mapstructure:"env" json:"env"`

	// URL is the endpoint for http/sse transport.
	URL string `mapstructure:"url" json:"url"`

	// Headers are HTTP headers for http/sse transport (supports ${VAR} expansion).
	Headers map[string]string `mapstructure:"headers" json:"headers"`

	// Enabled controls whether this server is active (default: true when nil).
	Enabled *bool `mapstructure:"enabled" json:"enabled"`

	// Timeout overrides the global default timeout for this server.
	Timeout time.Duration `mapstructure:"timeout" json:"timeout"`

	// SafetyLevel for tools from this server: "safe", "moderate", "dangerous" (default: "dangerous")
	SafetyLevel string `mapstructure:"safetyLevel" json:"safetyLevel"`
}

// IsEnabled returns whether the server is enabled (defaults to true when nil).
func (s MCPServerConfig) IsEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}
