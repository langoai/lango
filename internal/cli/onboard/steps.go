package onboard

import (
	"fmt"
	"sort"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/langoai/lango/internal/cli/settings"
	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

// NewProviderStepForm creates the Step 1 form: Provider Setup.
func NewProviderStepForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Provider Setup")

	form.AddField(&tuicore.Field{
		Key: "type", Label: "Provider Type", Type: tuicore.InputSelect,
		Value:       providerTypeFromConfig(cfg),
		Options:     []string{"anthropic", "openai", "gemini", "ollama", "github"},
		Description: "LLM provider type; determines API format and authentication method",
	})

	form.AddField(&tuicore.Field{
		Key: "id", Label: "Provider Name", Type: tuicore.InputText,
		Value:       providerIDFromConfig(cfg),
		Placeholder: "e.g. anthropic, my-openai",
		Description: "Unique identifier to reference this provider in other settings",
	})

	form.AddField(&tuicore.Field{
		Key: "apikey", Label: "API Key", Type: tuicore.InputPassword,
		Value:       providerAPIKeyFromConfig(cfg),
		Placeholder: "sk-... or ${ENV_VAR}",
		Description: "API key or environment variable reference for authentication",
	})

	form.AddField(&tuicore.Field{
		Key: "baseurl", Label: "Base URL (optional)", Type: tuicore.InputText,
		Value:       providerBaseURLFromConfig(cfg),
		Placeholder: "https://api.example.com/v1",
		Description: "Custom API endpoint; leave empty for provider default",
	})

	return &form
}

// NewAgentStepForm creates the Step 2 form: Agent Config.
func NewAgentStepForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Agent Config")

	providerOpts := buildProviderOptions(cfg)
	providerField := &tuicore.Field{
		Key: "provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:       cfg.Agent.Provider,
		Options:     providerOpts,
		Description: "LLM provider to use for agent inference",
	}
	form.AddField(providerField)

	modelField := &tuicore.Field{
		Key: "model", Label: "Model ID", Type: tuicore.InputSearchSelect,
		Value:       cfg.Agent.Model,
		Placeholder: suggestModel(cfg.Agent.Provider),
		Options:     []string{},
		Description: "Model identifier from the selected provider",
	}
	form.AddField(modelField)

	// Try to fetch models dynamically from the selected provider
	if modelOpts, fetchErr := settings.FetchModelOptionsWithError(cfg.Agent.Provider, cfg, cfg.Agent.Model); len(modelOpts) > 0 {
		modelField.Options = modelOpts
		modelField.Placeholder = ""
		modelField.Description = fmt.Sprintf("Fetched %d models from provider; press Enter to search", len(modelOpts))
	} else if fetchErr != nil {
		modelField.Type = tuicore.InputText
		modelField.Description = fmt.Sprintf("Could not fetch models (%v); enter model ID manually", fetchErr)
	} else {
		modelField.Type = tuicore.InputText
	}

	// Reactive: when provider changes, re-fetch model list and update placeholder
	cfgCopy := cfg
	providerField.OnChange = func(newProvider string) tea.Cmd {
		modelField.Placeholder = suggestModel(newProvider)
		modelField.Loading = true
		modelField.LoadError = nil
		return settings.FetchModelOptionsCmd("model", newProvider, cfgCopy, "")
	}

	form.AddField(&tuicore.Field{
		Key: "maxtokens", Label: "Max Tokens", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Agent.MaxTokens),
		Description: "Maximum number of tokens the model can generate per response",
		Validate: func(s string) error {
			v, err := strconv.Atoi(s)
			if err != nil {
				return fmt.Errorf("must be integer")
			}
			if v <= 0 {
				return fmt.Errorf("must be positive")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "temp", Label: "Temperature", Type: tuicore.InputText,
		Value:       fmt.Sprintf("%.1f", cfg.Agent.Temperature),
		Description: "Sampling temperature (0.0 = deterministic, 2.0 = max randomness)",
		Validate: func(s string) error {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("must be a number")
			}
			if v < 0.0 || v > 2.0 {
				return fmt.Errorf("must be between 0.0 and 2.0")
			}
			return nil
		},
	})

	return &form
}

// NewChannelStepForm creates the Step 3 form for the given channel type.
func NewChannelStepForm(channel string, cfg *config.Config) *tuicore.FormModel {
	switch types.ChannelType(channel) {
	case types.ChannelTelegram:
		return newTelegramForm(cfg)
	case types.ChannelDiscord:
		return newDiscordForm(cfg)
	case types.ChannelSlack:
		return newSlackForm(cfg)
	default:
		return nil
	}
}

func newTelegramForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Telegram Setup")
	form.AddField(&tuicore.Field{
		Key: "telegram_token", Label: "Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Telegram.BotToken,
		Placeholder: "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		Description: "Telegram bot token from @BotFather",
	})
	return &form
}

func newDiscordForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Discord Setup")
	form.AddField(&tuicore.Field{
		Key: "discord_token", Label: "Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Discord.BotToken,
		Placeholder: "Bot token from Developer Portal",
		Description: "Discord bot token from the Developer Portal",
	})
	return &form
}

func newSlackForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Slack Setup")
	form.AddField(&tuicore.Field{
		Key: "slack_token", Label: "Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.BotToken,
		Placeholder: "xoxb-...",
		Description: "Slack bot token (xoxb-...) from your Slack app",
	})
	form.AddField(&tuicore.Field{
		Key: "slack_app_token", Label: "App Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.AppToken,
		Placeholder: "xapp-...",
		Description: "Slack app-level token (xapp-...) for Socket Mode",
	})
	return &form
}

// NewSecurityStepForm creates the Step 4 form: Security & Auth.
func NewSecurityStepForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security & Auth")

	interceptorEnabled := &tuicore.Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.Enabled,
		Description: "Enable privacy interceptor to filter sensitive content",
	}
	form.AddField(interceptorEnabled)

	form.AddField(&tuicore.Field{
		Key: "interceptor_pii", Label: "  Redact PII", Type: tuicore.InputBool,
		Checked:     cfg.Security.Interceptor.RedactPII,
		Description: "Automatically redact personally identifiable information",
		VisibleWhen: func() bool { return interceptorEnabled.Checked },
	})

	policyVal := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policyVal == "" {
		policyVal = "dangerous"
	}
	form.AddField(&tuicore.Field{
		Key: "interceptor_policy", Label: "  Approval Policy", Type: tuicore.InputSelect,
		Value:       policyVal,
		Options:     []string{"dangerous", "all", "configured", "none"},
		Description: "Which tool calls require explicit user approval before execution",
		VisibleWhen: func() bool { return interceptorEnabled.Checked },
	})

	return &form
}

// buildProviderOptions builds a list of provider IDs from config.
func buildProviderOptions(cfg *config.Config) []string {
	opts := make([]string, 0, len(cfg.Providers))
	for id := range cfg.Providers {
		opts = append(opts, id)
	}
	sort.Strings(opts)
	if len(opts) == 0 {
		opts = []string{"anthropic", "openai", "gemini", "ollama", "github"}
	}
	return opts
}

// suggestModel returns a default model suggestion for the given provider.
func suggestModel(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-sonnet-4-5-20250929"
	case "openai":
		return "gpt-4o"
	case "gemini":
		return "gemini-2.0-flash"
	case "ollama":
		return "llama3.1"
	case "github":
		return "gpt-4o"
	default:
		return "claude-sonnet-4-5-20250929"
	}
}

// providerTypeFromConfig extracts the provider type from the first provider.
func providerTypeFromConfig(cfg *config.Config) string {
	if len(cfg.Providers) == 0 {
		return "anthropic"
	}
	// Use the agent's provider if available
	if p, ok := cfg.Providers[cfg.Agent.Provider]; ok {
		return string(p.Type)
	}
	// Otherwise use first provider
	for _, p := range cfg.Providers {
		return string(p.Type)
	}
	return "anthropic"
}

// providerIDFromConfig extracts the provider ID from config.
func providerIDFromConfig(cfg *config.Config) string {
	if cfg.Agent.Provider != "" {
		return cfg.Agent.Provider
	}
	for id := range cfg.Providers {
		return id
	}
	return "anthropic"
}

// providerAPIKeyFromConfig extracts the API key from the primary provider.
func providerAPIKeyFromConfig(cfg *config.Config) string {
	if p, ok := cfg.Providers[cfg.Agent.Provider]; ok {
		return p.APIKey
	}
	for _, p := range cfg.Providers {
		return p.APIKey
	}
	return ""
}

// providerBaseURLFromConfig extracts the base URL from the primary provider.
func providerBaseURLFromConfig(cfg *config.Config) string {
	if p, ok := cfg.Providers[cfg.Agent.Provider]; ok {
		return p.BaseURL
	}
	for _, p := range cfg.Providers {
		return p.BaseURL
	}
	return ""
}
