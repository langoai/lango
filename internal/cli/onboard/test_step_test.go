package onboard

import (
	"testing"

	"github.com/langowarny/lango/internal/config"
)

func TestRunConfigTests_ValidConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-test-key"},
	}
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"

	results := RunConfigTests(cfg)

	for _, r := range results {
		if r.Status == "fail" {
			t.Errorf("unexpected fail for %q: %s", r.Name, r.Message)
		}
	}

	// Provider and agent model should pass
	provResult := findResult(results, "Provider configured")
	if provResult == nil || provResult.Status != "pass" {
		t.Error("Provider configured should pass")
	}

	modelResult := findResult(results, "Agent model set")
	if modelResult == nil || modelResult.Status != "pass" {
		t.Error("Agent model set should pass")
	}
}

func TestRunConfigTests_MissingProvider(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Agent.Provider = "nonexistent"
	cfg.Agent.Model = "some-model"

	results := RunConfigTests(cfg)

	provResult := findResult(results, "Provider configured")
	if provResult == nil || provResult.Status != "fail" {
		t.Error("Provider configured should fail")
	}
}

func TestRunConfigTests_MissingAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: ""},
	}
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"

	results := RunConfigTests(cfg)

	keyResult := findResult(results, "API key set")
	if keyResult == nil || keyResult.Status != "fail" {
		t.Error("API key set should fail when empty")
	}
}

func TestRunConfigTests_PlaceholderAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-..."},
	}
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"

	results := RunConfigTests(cfg)

	keyResult := findResult(results, "API key set")
	if keyResult == nil || keyResult.Status != "warn" {
		t.Error("API key set should warn for placeholder")
	}
}

func TestRunConfigTests_OllamaNoAPIKey(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"ollama": {Type: "ollama"},
	}
	cfg.Agent.Provider = "ollama"
	cfg.Agent.Model = "llama3.1"

	results := RunConfigTests(cfg)

	keyResult := findResult(results, "API key set")
	if keyResult == nil || keyResult.Status != "pass" {
		t.Error("API key set should pass for ollama (no key required)")
	}
}

func TestRunConfigTests_DisabledChannel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-real-key"},
	}
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	// No channels enabled

	results := RunConfigTests(cfg)

	chResult := findResult(results, "Channel tokens")
	if chResult == nil || chResult.Status != "warn" {
		t.Error("Channel tokens should warn when no channels enabled")
	}
}

func TestRunConfigTests_EnabledChannelMissingToken(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-real-key"},
	}
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = "claude-sonnet-4-5-20250929"
	cfg.Channels.Telegram.Enabled = true
	cfg.Channels.Telegram.BotToken = ""

	results := RunConfigTests(cfg)

	chResult := findResult(results, "Channel tokens")
	if chResult == nil || chResult.Status != "fail" {
		t.Error("Channel tokens should fail when telegram enabled but token empty")
	}
}

func TestRunConfigTests_MissingModel(t *testing.T) {
	cfg := config.DefaultConfig()
	cfg.Providers = map[string]config.ProviderConfig{
		"anthropic": {Type: "anthropic", APIKey: "sk-real-key"},
	}
	cfg.Agent.Provider = "anthropic"
	cfg.Agent.Model = ""

	results := RunConfigTests(cfg)

	modelResult := findResult(results, "Agent model set")
	if modelResult == nil || modelResult.Status != "fail" {
		t.Error("Agent model set should fail when empty")
	}
}

func findResult(results []TestResult, name string) *TestResult {
	for i := range results {
		if results[i].Name == name {
			return &results[i]
		}
	}
	return nil
}
