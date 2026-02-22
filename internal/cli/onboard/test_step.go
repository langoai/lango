package onboard

import (
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// TestResult represents a single configuration validation result.
type TestResult struct {
	Name    string
	Status  string // "pass", "warn", "fail"
	Message string
}

// RunConfigTests validates the configuration and returns test results.
func RunConfigTests(cfg *config.Config) []TestResult {
	var results []TestResult

	// 1. Provider exists
	results = append(results, checkProviderExists(cfg))

	// 2. API key is set
	results = append(results, checkAPIKey(cfg))

	// 3. Agent model is set
	results = append(results, checkAgentModel(cfg))

	// 4. Channel token (if any channel is enabled)
	results = append(results, checkChannelTokens(cfg))

	// 5. Config validates
	results = append(results, checkConfigValidate(cfg))

	return results
}

func checkProviderExists(cfg *config.Config) TestResult {
	providerID := cfg.Agent.Provider
	if providerID == "" {
		return TestResult{
			Name:    "Provider configured",
			Status:  "fail",
			Message: "No provider selected in agent config",
		}
	}

	p, ok := cfg.Providers[providerID]
	if !ok {
		return TestResult{
			Name:    "Provider configured",
			Status:  "fail",
			Message: fmt.Sprintf("Provider %q not found in providers map", providerID),
		}
	}

	if p.Type == "" {
		return TestResult{
			Name:    "Provider configured",
			Status:  "fail",
			Message: fmt.Sprintf("Provider %q has no type", providerID),
		}
	}

	return TestResult{
		Name:    "Provider configured",
		Status:  "pass",
		Message: fmt.Sprintf("Provider %q (%s) is configured", providerID, p.Type),
	}
}

func checkAPIKey(cfg *config.Config) TestResult {
	providerID := cfg.Agent.Provider
	if providerID == "" {
		return TestResult{
			Name:    "API key set",
			Status:  "fail",
			Message: "No provider selected",
		}
	}

	p, ok := cfg.Providers[providerID]
	if !ok {
		return TestResult{
			Name:    "API key set",
			Status:  "fail",
			Message: "Provider not found",
		}
	}

	// Ollama doesn't need an API key
	if p.Type == "ollama" {
		return TestResult{
			Name:    "API key set",
			Status:  "pass",
			Message: "Ollama does not require an API key",
		}
	}

	if p.APIKey == "" {
		return TestResult{
			Name:    "API key set",
			Status:  "fail",
			Message: "API key is empty",
		}
	}

	if p.APIKey == "sk-..." || p.APIKey == "placeholder" {
		return TestResult{
			Name:    "API key set",
			Status:  "warn",
			Message: "API key looks like a placeholder",
		}
	}

	return TestResult{
		Name:    "API key set",
		Status:  "pass",
		Message: "API key is configured",
	}
}

func checkAgentModel(cfg *config.Config) TestResult {
	if cfg.Agent.Model == "" {
		return TestResult{
			Name:    "Agent model set",
			Status:  "fail",
			Message: "No model specified in agent config",
		}
	}

	return TestResult{
		Name:    "Agent model set",
		Status:  "pass",
		Message: fmt.Sprintf("Model: %s", cfg.Agent.Model),
	}
}

func checkChannelTokens(cfg *config.Config) TestResult {
	anyEnabled := false

	if cfg.Channels.Telegram.Enabled {
		anyEnabled = true
		if cfg.Channels.Telegram.BotToken == "" {
			return TestResult{
				Name:    "Channel tokens",
				Status:  "fail",
				Message: "Telegram is enabled but bot token is empty",
			}
		}
	}

	if cfg.Channels.Discord.Enabled {
		anyEnabled = true
		if cfg.Channels.Discord.BotToken == "" {
			return TestResult{
				Name:    "Channel tokens",
				Status:  "fail",
				Message: "Discord is enabled but bot token is empty",
			}
		}
	}

	if cfg.Channels.Slack.Enabled {
		anyEnabled = true
		if cfg.Channels.Slack.BotToken == "" {
			return TestResult{
				Name:    "Channel tokens",
				Status:  "fail",
				Message: "Slack is enabled but bot token is empty",
			}
		}
	}

	if !anyEnabled {
		return TestResult{
			Name:    "Channel tokens",
			Status:  "warn",
			Message: "No channels enabled (configure later in settings)",
		}
	}

	return TestResult{
		Name:    "Channel tokens",
		Status:  "pass",
		Message: "Enabled channel tokens are present",
	}
}

func checkConfigValidate(cfg *config.Config) TestResult {
	if err := config.Validate(cfg); err != nil {
		return TestResult{
			Name:    "Config validation",
			Status:  "fail",
			Message: err.Error(),
		}
	}

	return TestResult{
		Name:    "Config validation",
		Status:  "pass",
		Message: "Configuration is valid",
	}
}
