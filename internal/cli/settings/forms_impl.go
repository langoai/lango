package settings

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// buildProviderOptions builds provider options from registered providers.
func buildProviderOptions(cfg *config.Config) []string {
	opts := make([]string, 0, len(cfg.Providers))
	for id := range cfg.Providers {
		opts = append(opts, id)
	}
	sort.Strings(opts)
	if len(opts) == 0 {
		opts = []string{"anthropic", "openai", "gemini", "ollama"}
	}
	return opts
}

// NewAgentForm creates the Agent configuration form.
func NewAgentForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Agent Configuration")

	providerOpts := buildProviderOptions(cfg)
	providerField := &tuicore.Field{
		Key: "provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.Agent.Provider,
		Options:     providerOpts,
		Description: "LLM provider to use for agent inference",
	}
	form.AddField(providerField)

	modelField := &tuicore.Field{
		Key: "model", Label: "Model ID", Type: tuicore.InputText,
		Value:       cfg.Agent.Model,
		Placeholder: "e.g. claude-3-5-sonnet-20240620",
		Description: "Model identifier from the selected provider",
	}
	form.AddField(modelField)

	// Try to fetch models dynamically from the selected provider
	if modelOpts, fetchErr := FetchModelOptionsWithError(cfg.Agent.Provider, cfg, cfg.Agent.Model); len(modelOpts) > 0 {
		modelField.Type = tuicore.InputSearchSelect
		modelField.Options = modelOpts
		modelField.Placeholder = ""
		modelField.Description = fmt.Sprintf("Fetched %d models from provider; press Enter to search", len(modelOpts))
	} else if fetchErr != nil {
		modelField.Description = fmt.Sprintf("Could not fetch models (%v); enter model ID manually", fetchErr)
	}

	// Reactive: when provider changes, re-fetch model list
	cfgCopy := cfg
	providerField.OnChange = func(newProvider string) tea.Cmd {
		modelField.Loading = true
		modelField.LoadError = nil
		return FetchModelOptionsCmd("model", newProvider, cfgCopy, "")
	}

	form.AddField(&tuicore.Field{
		Key: "maxtokens", Label: "Max Tokens", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Agent.MaxTokens),
		Description: "Maximum number of tokens the model can generate per response",
		Validate: func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "temp", Label: "Temperature", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.Agent.Temperature),
		Placeholder: "0.0 to 2.0",
		Description: "Controls randomness: 0.0 = deterministic, 2.0 = maximum creativity",
		Validate: func(s string) error {
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if f < 0 || f > 2.0 {
				return fmt.Errorf("must be between 0.0 and 2.0")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "prompts_dir", Label: "Prompts Directory", Type: tuicore.InputText,
		Value:       cfg.Agent.PromptsDir,
		Placeholder: "~/.lango/prompts (supports agents/<name>/ for per-agent overrides)",
		Description: "Directory containing system prompt templates; supports per-agent overrides",
	})

	fallbackOpts := append([]string{""}, providerOpts...)
	fbProviderField := &tuicore.Field{
		Key: "fallback_provider", Label: "Fallback Provider", Type: tuicore.InputSelect,
		Value:       cfg.Agent.FallbackProvider,
		Options:     fallbackOpts,
		Description: "Alternative provider used when primary provider fails or is unavailable",
	}
	form.AddField(fbProviderField)

	fbModelField := &tuicore.Field{
		Key: "fallback_model", Label: "Fallback Model", Type: tuicore.InputText,
		Value:       cfg.Agent.FallbackModel,
		Placeholder: "e.g. gpt-4o",
		Description: "Model to use with the fallback provider",
	}
	form.AddField(fbModelField)

	if cfg.Agent.FallbackProvider != "" {
		if fbModelOpts, fbErr := FetchModelOptionsWithError(cfg.Agent.FallbackProvider, cfg, cfg.Agent.FallbackModel); len(fbModelOpts) > 0 {
			fbModelOpts = append([]string{""}, fbModelOpts...)
			fbModelField.Type = tuicore.InputSearchSelect
			fbModelField.Options = fbModelOpts
			fbModelField.Placeholder = ""
		} else if fbErr != nil {
			fbModelField.Description = fmt.Sprintf("Could not fetch models (%v); enter model ID manually", fbErr)
		}
	}

	// Reactive: when fallback provider changes, re-fetch fallback model list
	fbProviderField.OnChange = func(newProvider string) tea.Cmd {
		if newProvider == "" {
			return nil
		}
		fbModelField.Loading = true
		fbModelField.LoadError = nil
		return FetchModelOptionsCmd("fallback_model", newProvider, cfgCopy, "")
	}

	form.AddField(&tuicore.Field{
		Key: "request_timeout", Label: "Request Timeout", Type: tuicore.InputText,
		Value:       cfg.Agent.RequestTimeout.String(),
		Placeholder: "5m (e.g. 30s, 2m, 5m)",
		Description: "Maximum time to wait for a single LLM API request",
	})

	form.AddField(&tuicore.Field{
		Key: "tool_timeout", Label: "Tool Timeout", Type: tuicore.InputText,
		Value:       cfg.Agent.ToolTimeout.String(),
		Placeholder: "2m (e.g. 30s, 1m, 2m)",
		Description: "Maximum execution time allowed for a single tool invocation",
	})

	return &form
}

// NewServerForm creates the Server configuration form.
func NewServerForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Server Configuration")

	form.AddField(&tuicore.Field{
		Key: "host", Label: "Host", Type: tuicore.InputText,
		Value:       cfg.Server.Host,
		Description: "Network interface to bind to; use 0.0.0.0 to listen on all interfaces",
	})

	form.AddField(&tuicore.Field{
		Key: "port", Label: "Port", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Server.Port),
		Validate:    validatePort,
		Description: "TCP port for the HTTP/WebSocket server (1-65535)",
	})

	form.AddField(&tuicore.Field{
		Key: "http", Label: "Generic HTTP", Type: tuicore.InputBool,
		Checked:     cfg.Server.HTTPEnabled,
		Description: "Enable REST API endpoint for HTTP-based integrations",
	})

	form.AddField(&tuicore.Field{
		Key: "ws", Label: "WebSockets", Type: tuicore.InputBool,
		Checked:     cfg.Server.WebSocketEnabled,
		Description: "Enable WebSocket endpoint for real-time bidirectional communication",
	})

	return &form
}

// NewChannelsForm creates the Channels configuration form.
func NewChannelsForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Channels Configuration")

	telegramEnabled := &tuicore.Field{
		Key: "telegram_enabled", Label: "Telegram", Type: tuicore.InputBool,
		Checked:     cfg.Channels.Telegram.Enabled,
		Description: "Enable Telegram bot channel for receiving and sending messages",
	}
	form.AddField(telegramEnabled)
	form.AddField(&tuicore.Field{
		Key: "telegram_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Telegram.BotToken,
		Placeholder: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		Description: "Bot token from @BotFather; use ${ENV_VAR} to reference environment variables",
		VisibleWhen: func() bool { return telegramEnabled.Checked },
	})

	discordEnabled := &tuicore.Field{
		Key: "discord_enabled", Label: "Discord", Type: tuicore.InputBool,
		Checked:     cfg.Channels.Discord.Enabled,
		Description: "Enable Discord bot channel for receiving and sending messages",
	}
	form.AddField(discordEnabled)
	form.AddField(&tuicore.Field{
		Key: "discord_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Discord.BotToken,
		Description: "Bot token from Discord Developer Portal; use ${ENV_VAR} for security",
		VisibleWhen: func() bool { return discordEnabled.Checked },
	})

	slackEnabled := &tuicore.Field{
		Key: "slack_enabled", Label: "Slack", Type: tuicore.InputBool,
		Checked:     cfg.Channels.Slack.Enabled,
		Description: "Enable Slack bot channel using Socket Mode",
	}
	form.AddField(slackEnabled)
	form.AddField(&tuicore.Field{
		Key: "slack_token", Label: "  Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.BotToken,
		Placeholder: "xoxb-...",
		Description: "Slack Bot User OAuth Token (starts with xoxb-)",
		VisibleWhen: func() bool { return slackEnabled.Checked },
	})
	form.AddField(&tuicore.Field{
		Key: "slack_app_token", Label: "  App Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.AppToken,
		Placeholder: "xapp-...",
		Description: "Slack App-Level Token for Socket Mode (starts with xapp-)",
		VisibleWhen: func() bool { return slackEnabled.Checked },
	})

	return &form
}

// NewToolsForm creates the Tools configuration form.
func NewToolsForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Tools Configuration")

	form.AddField(&tuicore.Field{
		Key: "exec_timeout", Label: "Exec Timeout", Type: tuicore.InputText,
		Value:       cfg.Tools.Exec.DefaultTimeout.String(),
		Placeholder: "30s",
		Description: "Default timeout for shell command execution",
	})
	form.AddField(&tuicore.Field{
		Key: "exec_bg", Label: "Allow Background", Type: tuicore.InputBool,
		Checked:     cfg.Tools.Exec.AllowBackground,
		Description: "Allow the agent to run shell commands in the background",
	})

	form.AddField(&tuicore.Field{
		Key: "browser_enabled", Label: "Browser Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Tools.Browser.Enabled,
		Description: "Enable headless browser tool for web scraping and interaction",
	})
	form.AddField(&tuicore.Field{
		Key: "browser_headless", Label: "Browser Headless", Type: tuicore.InputBool,
		Checked:     cfg.Tools.Browser.Headless,
		Description: "Run browser without visible UI window; disable for debugging",
	})
	form.AddField(&tuicore.Field{
		Key: "browser_session_timeout", Label: "Browser Session Timeout", Type: tuicore.InputText,
		Value:       cfg.Tools.Browser.SessionTimeout.String(),
		Placeholder: "5m",
		Description: "Maximum duration for a single browser session before auto-close",
	})

	form.AddField(&tuicore.Field{
		Key: "fs_max_read", Label: "Max Read Size", Type: tuicore.InputInt,
		Value:       strconv.FormatInt(cfg.Tools.Filesystem.MaxReadSize, 10),
		Description: "Maximum file size in bytes that the filesystem tool can read",
		Validate: func(s string) error {
			if i, err := strconv.ParseInt(s, 10, 64); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// NewSessionForm creates the Session configuration form.
func NewSessionForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Session Configuration")

	form.AddField(&tuicore.Field{
		Key: "ttl", Label: "Session TTL", Type: tuicore.InputText,
		Value:       cfg.Session.TTL.String(),
		Description: "Time-to-live before an idle session expires (e.g. 24h, 7d)",
	})

	form.AddField(&tuicore.Field{
		Key: "max_history_turns", Label: "Max History Turns", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Session.MaxHistoryTurns),
		Description: "Maximum number of conversation turns kept in session history",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	return &form
}

// formatCustomPatterns formats a map of custom patterns into a comma-separated
// "name:regex" string for display in the TUI.
func formatCustomPatterns(patterns map[string]string) string {
	if len(patterns) == 0 {
		return ""
	}
	parts := make([]string, 0, len(patterns))
	for name, regex := range patterns {
		parts = append(parts, name+":"+regex)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

// ParseCustomPatterns parses a comma-separated "name:regex" string into a map.
func ParseCustomPatterns(val string) map[string]string {
	if val == "" {
		return nil
	}
	result := make(map[string]string)
	parts := strings.Split(val, ",")
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		idx := strings.Index(p, ":")
		if idx <= 0 || idx >= len(p)-1 {
			continue
		}
		name := strings.TrimSpace(p[:idx])
		regex := strings.TrimSpace(p[idx+1:])
		if name != "" && regex != "" {
			result[name] = regex
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

func validatePort(s string) error {
	p, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid number")
	}
	if p < 1 || p > 65535 {
		return fmt.Errorf("port out of range")
	}
	return nil
}

// NewOIDCProviderForm creates the OIDC Provider configuration form.
func NewOIDCProviderForm(id string, cfg config.OIDCProviderConfig) *tuicore.FormModel {
	title := "Edit OIDC Provider: " + id
	if id == "" {
		title = "Add New OIDC Provider"
	}
	form := tuicore.NewFormModel(title)

	if id == "" {
		form.AddField(&tuicore.Field{
			Key: "oidc_id", Label: "Provider Name", Type: tuicore.InputText,
			Placeholder: "e.g. google, github",
			Description: "Unique identifier for this OIDC provider configuration",
		})
	}

	form.AddField(&tuicore.Field{
		Key: "oidc_issuer", Label: "Issuer URL", Type: tuicore.InputText,
		Value:       cfg.IssuerURL,
		Placeholder: "https://accounts.google.com",
		Description: "OIDC issuer URL used for auto-discovery of endpoints",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_client_id", Label: "Client ID", Type: tuicore.InputPassword,
		Value:       cfg.ClientID,
		Placeholder: "${ENV_VAR} or value",
		Description: "OAuth2 Client ID from the identity provider; supports ${ENV_VAR} syntax",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_client_secret", Label: "Client Secret", Type: tuicore.InputPassword,
		Value:       cfg.ClientSecret,
		Placeholder: "${ENV_VAR} or value",
		Description: "OAuth2 Client Secret; strongly recommend using ${ENV_VAR} for security",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_redirect", Label: "Redirect URL", Type: tuicore.InputText,
		Value:       cfg.RedirectURL,
		Placeholder: "http://localhost:18789/auth/callback/<name>",
		Description: "Callback URL registered with the identity provider",
	})

	form.AddField(&tuicore.Field{
		Key: "oidc_scopes", Label: "Scopes", Type: tuicore.InputText,
		Value:       strings.Join(cfg.Scopes, ","),
		Placeholder: "openid,email,profile",
		Description: "OAuth2 scopes to request; openid is required for OIDC",
	})

	return &form
}

// derefBool safely dereferences a *bool with a default value.
func derefBool(p *bool, def bool) bool {
	if p == nil {
		return def
	}
	return *p
}

// formatKeyValueMap formats a map[string]string as "key:value" comma-separated.
func formatKeyValueMap(m map[string]string) string {
	return formatCustomPatterns(m)
}

// NewProviderForm creates a Provider configuration form.
func NewProviderForm(id string, cfg config.ProviderConfig) *tuicore.FormModel {
	title := "Edit Provider: " + id
	if id == "" {
		title = "Add New Provider"
	}
	form := tuicore.NewFormModel(title)

	form.AddField(&tuicore.Field{
		Key: "type", Label: "Type", Type: tuicore.InputSelect,
		Value:       string(cfg.Type),
		Options:     []string{"openai", "anthropic", "gemini", "ollama", "github"},
		Description: "LLM provider type; determines API format and authentication method",
	})

	if id == "" {
		form.AddField(&tuicore.Field{
			Key: "id", Label: "Provider Name", Type: tuicore.InputText,
			Placeholder: "e.g. my-openai, production-claude",
			Description: "Unique identifier to reference this provider in other settings",
		})
	}

	form.AddField(&tuicore.Field{
		Key: "apikey", Label: "API Key", Type: tuicore.InputPassword,
		Value:       cfg.APIKey,
		Placeholder: "${ENV_VAR} or key",
		Description: "API key for authentication; use ${ENV_VAR} to reference environment variables",
	})

	form.AddField(&tuicore.Field{
		Key: "baseurl", Label: "Base URL", Type: tuicore.InputText,
		Value:       cfg.BaseURL,
		Placeholder: "https://api.example.com/v1",
		Description: "Custom API base URL; leave empty for provider's default endpoint",
	})

	return &form
}
