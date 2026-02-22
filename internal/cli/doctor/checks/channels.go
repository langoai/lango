package checks

import (
	"context"
	"os"

	"github.com/langoai/lango/internal/config"
)

// ChannelCheck validates channel token configurations.
type ChannelCheck struct{}

// Name returns the check name.
func (c *ChannelCheck) Name() string {
	return "Channel Tokens"
}

// Run checks if enabled channels have valid tokens configured.
func (c *ChannelCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "No configuration loaded",
		}
	}

	var issues []string
	var configured []string

	// Check Telegram
	if cfg.Channels.Telegram.Enabled {
		token := resolveEnvValue(cfg.Channels.Telegram.BotToken)
		if token == "" {
			issues = append(issues, "Telegram: bot token not set")
		} else {
			configured = append(configured, "Telegram")
		}
	}

	// Check Discord
	if cfg.Channels.Discord.Enabled {
		token := resolveEnvValue(cfg.Channels.Discord.BotToken)
		if token == "" {
			issues = append(issues, "Discord: bot token not set")
		} else {
			configured = append(configured, "Discord")
		}
	}

	// Check Slack
	if cfg.Channels.Slack.Enabled {
		botToken := resolveEnvValue(cfg.Channels.Slack.BotToken)
		appToken := resolveEnvValue(cfg.Channels.Slack.AppToken)
		if botToken == "" {
			issues = append(issues, "Slack: bot token not set")
		} else if appToken == "" {
			issues = append(issues, "Slack: app token not set")
		} else {
			configured = append(configured, "Slack")
		}
	}

	// No channels enabled
	if !cfg.Channels.Telegram.Enabled && !cfg.Channels.Discord.Enabled && !cfg.Channels.Slack.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "No channels enabled",
			Details: "Enable at least one channel in configuration or run 'lango onboard'",
		}
	}

	if len(issues) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusFail,
			Message: "Channel token issues found",
			Details: joinStrings(issues, "; "),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Channel tokens configured",
		Details: joinStrings(configured, ", "),
	}
}

// Fix cannot auto-fix channel tokens - requires user input.
func (c *ChannelCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return Result{
		Name:    c.Name(),
		Status:  StatusSkip,
		Message: "Channel tokens require manual configuration",
		Details: "Run 'lango onboard' to configure channels",
	}
}

// resolveEnvValue resolves ${VAR} patterns to environment variable values.
func resolveEnvValue(value string) string {
	if len(value) > 3 && value[0:2] == "${" && value[len(value)-1] == '}' {
		envVar := value[2 : len(value)-1]
		return os.Getenv(envVar)
	}
	return value
}

// joinStrings joins strings with a separator.
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
