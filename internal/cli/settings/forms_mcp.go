package settings

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewMCPForm creates the MCP Servers configuration form.
func NewMCPForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("MCP Servers Configuration")

	form.AddField(&tuicore.Field{
		Key: "mcp_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.MCP.Enabled,
		Description: "Enable MCP server integration",
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_default_timeout", Label: "Default Timeout", Type: tuicore.InputText,
		Value:       cfg.MCP.DefaultTimeout.String(),
		Placeholder: "30s",
		Description: "Default timeout for MCP operations (e.g. 30s, 1m)",
		Validate: func(s string) error {
			if _, err := time.ParseDuration(s); err != nil {
				return fmt.Errorf("invalid duration: %s", s)
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_max_output_tokens", Label: "Max Output Tokens", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.MCP.MaxOutputTokens),
		Description: "Maximum output tokens from MCP tool calls",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_health_check_interval", Label: "Health Check Interval", Type: tuicore.InputText,
		Value:       cfg.MCP.HealthCheckInterval.String(),
		Placeholder: "30s",
		Description: "Interval for periodic server health probes",
		Validate: func(s string) error {
			if _, err := time.ParseDuration(s); err != nil {
				return fmt.Errorf("invalid duration: %s", s)
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_auto_reconnect", Label: "Auto Reconnect", Type: tuicore.InputBool,
		Checked:     cfg.MCP.AutoReconnect,
		Description: "Automatically reconnect on connection loss",
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_max_reconnect_attempts", Label: "Max Reconnect Attempts", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.MCP.MaxReconnectAttempts),
		Description: "Maximum number of reconnection attempts",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewMCPServerForm creates a form for adding or editing a single MCP server.
func NewMCPServerForm(name string, srv config.MCPServerConfig) *tuicore.FormModel {
	title := "Edit MCP Server: " + name
	if name == "" {
		title = "Add New MCP Server"
	}
	form := tuicore.NewFormModel(title)

	if name == "" {
		form.AddField(&tuicore.Field{
			Key: "mcp_srv_name", Label: "Server Name", Type: tuicore.InputText,
			Placeholder: "e.g. github, filesystem",
			Description: "Unique name to identify this MCP server",
		})
	}

	transport := srv.Transport
	if transport == "" {
		transport = "stdio"
	}

	transportField := &tuicore.Field{
		Key: "mcp_srv_transport", Label: "Transport", Type: tuicore.InputSelect,
		Value:       transport,
		Options:     []string{"stdio", "http", "sse"},
		Description: "Connection transport: stdio (subprocess), http (streamable), sse (server-sent events)",
	}
	form.AddField(transportField)

	commandField := &tuicore.Field{
		Key: "mcp_srv_command", Label: "Command", Type: tuicore.InputText,
		Value:       srv.Command,
		Placeholder: "e.g. npx, uvx, node",
		Description: "Executable to run for stdio transport",
		VisibleWhen: func() bool { return transportField.Value == "stdio" },
	}
	form.AddField(commandField)

	form.AddField(&tuicore.Field{
		Key: "mcp_srv_args", Label: "Args", Type: tuicore.InputText,
		Value:       strings.Join(srv.Args, ","),
		Placeholder: "arg1,arg2,arg3",
		Description: "Command arguments (comma-separated) for stdio transport",
		VisibleWhen: func() bool { return transportField.Value == "stdio" },
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_srv_url", Label: "URL", Type: tuicore.InputText,
		Value:       srv.URL,
		Placeholder: "https://example.com/mcp",
		Description: "Server endpoint URL for http/sse transport",
		VisibleWhen: func() bool { return transportField.Value == "http" || transportField.Value == "sse" },
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_srv_env", Label: "Environment", Type: tuicore.InputText,
		Value:       formatKeyValuePairs(srv.Env),
		Placeholder: "KEY=VAL,KEY2=VAL2",
		Description: "Environment variables (KEY=VAL,KEY=VAL); supports ${VAR} expansion",
	})

	form.AddField(&tuicore.Field{
		Key: "mcp_srv_headers", Label: "Headers", Type: tuicore.InputText,
		Value:       formatKeyValuePairs(srv.Headers),
		Placeholder: "Authorization=Bearer ${TOKEN}",
		Description: "HTTP headers (KEY=VAL,KEY=VAL) for http/sse transport",
		VisibleWhen: func() bool { return transportField.Value == "http" || transportField.Value == "sse" },
	})

	enabled := srv.IsEnabled()
	form.AddField(&tuicore.Field{
		Key: "mcp_srv_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     enabled,
		Description: "Whether this server is active",
	})

	timeoutVal := ""
	if srv.Timeout > 0 {
		timeoutVal = srv.Timeout.String()
	}
	form.AddField(&tuicore.Field{
		Key: "mcp_srv_timeout", Label: "Timeout Override", Type: tuicore.InputText,
		Value:       timeoutVal,
		Placeholder: "30s (empty = use global default)",
		Description: "Per-server timeout override; leave empty for global default",
		Validate: func(s string) error {
			if s == "" {
				return nil
			}
			if _, err := time.ParseDuration(s); err != nil {
				return fmt.Errorf("invalid duration: %s", s)
			}
			return nil
		},
	})

	safetyLevel := srv.SafetyLevel
	if safetyLevel == "" {
		safetyLevel = "dangerous"
	}
	form.AddField(&tuicore.Field{
		Key: "mcp_srv_safety", Label: "Safety Level", Type: tuicore.InputSelect,
		Value:       safetyLevel,
		Options:     []string{"safe", "moderate", "dangerous"},
		Description: "Tool safety classification for approval policy",
	})

	return &form
}

// formatKeyValuePairs converts a map to "KEY=VAL,KEY=VAL" string.
func formatKeyValuePairs(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+m[k])
	}
	return strings.Join(pairs, ",")
}
