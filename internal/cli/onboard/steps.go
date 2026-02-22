package onboard

import (
	"fmt"
	"sort"
	"strconv"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/types"
)

// NewProviderStepForm creates the Step 1 form: Provider Setup.
func NewProviderStepForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Provider Setup")

	form.AddField(&tuicore.Field{
		Key: "type", Label: "Provider Type", Type: tuicore.InputSelect,
		Value:   providerTypeFromConfig(cfg),
		Options: []string{"anthropic", "openai", "gemini", "ollama"},
	})

	form.AddField(&tuicore.Field{
		Key: "id", Label: "Provider Name", Type: tuicore.InputText,
		Value:       providerIDFromConfig(cfg),
		Placeholder: "e.g. anthropic, my-openai",
	})

	form.AddField(&tuicore.Field{
		Key: "apikey", Label: "API Key", Type: tuicore.InputPassword,
		Value:       providerAPIKeyFromConfig(cfg),
		Placeholder: "sk-... or ${ENV_VAR}",
	})

	form.AddField(&tuicore.Field{
		Key: "baseurl", Label: "Base URL (optional)", Type: tuicore.InputText,
		Value:       providerBaseURLFromConfig(cfg),
		Placeholder: "https://api.example.com/v1",
	})

	return &form
}

// NewAgentStepForm creates the Step 2 form: Agent Config.
func NewAgentStepForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Agent Config")

	providerOpts := buildProviderOptions(cfg)
	form.AddField(&tuicore.Field{
		Key: "provider", Label: "Provider", Type: tuicore.InputSelect,
		Value:   cfg.Agent.Provider,
		Options: providerOpts,
	})

	form.AddField(&tuicore.Field{
		Key: "model", Label: "Model ID", Type: tuicore.InputText,
		Value:       cfg.Agent.Model,
		Placeholder: suggestModel(cfg.Agent.Provider),
	})

	form.AddField(&tuicore.Field{
		Key: "maxtokens", Label: "Max Tokens", Type: tuicore.InputInt,
		Value: strconv.Itoa(cfg.Agent.MaxTokens),
		Validate: func(s string) error {
			if _, err := strconv.Atoi(s); err != nil {
				return fmt.Errorf("must be integer")
			}
			return nil
		},
	})

	form.AddField(&tuicore.Field{
		Key: "temp", Label: "Temperature", Type: tuicore.InputText,
		Value: fmt.Sprintf("%.1f", cfg.Agent.Temperature),
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
	})
	return &form
}

func newDiscordForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Discord Setup")
	form.AddField(&tuicore.Field{
		Key: "discord_token", Label: "Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Discord.BotToken,
		Placeholder: "Bot token from Developer Portal",
	})
	return &form
}

func newSlackForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Slack Setup")
	form.AddField(&tuicore.Field{
		Key: "slack_token", Label: "Bot Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.BotToken,
		Placeholder: "xoxb-...",
	})
	form.AddField(&tuicore.Field{
		Key: "slack_app_token", Label: "App Token", Type: tuicore.InputPassword,
		Value:       cfg.Channels.Slack.AppToken,
		Placeholder: "xapp-...",
	})
	return &form
}

// NewSecurityStepForm creates the Step 4 form: Security & Auth.
func NewSecurityStepForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Security & Auth")

	form.AddField(&tuicore.Field{
		Key: "interceptor_enabled", Label: "Privacy Interceptor", Type: tuicore.InputBool,
		Checked: cfg.Security.Interceptor.Enabled,
	})

	form.AddField(&tuicore.Field{
		Key: "interceptor_pii", Label: "Redact PII", Type: tuicore.InputBool,
		Checked: cfg.Security.Interceptor.RedactPII,
	})

	policyVal := string(cfg.Security.Interceptor.ApprovalPolicy)
	if policyVal == "" {
		policyVal = "dangerous"
	}
	form.AddField(&tuicore.Field{
		Key: "interceptor_policy", Label: "Approval Policy", Type: tuicore.InputSelect,
		Value:   policyVal,
		Options: []string{"dangerous", "all", "configured", "none"},
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
		opts = []string{"anthropic", "openai", "gemini", "ollama"}
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
