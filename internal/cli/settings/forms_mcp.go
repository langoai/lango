package settings

import (
	"fmt"
	"strconv"
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
