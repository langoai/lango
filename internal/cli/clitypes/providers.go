package clitypes

import (
	"os"

	"github.com/langowarny/lango/internal/provider"
)

// ProviderMetadata holds information about a provider for CLI display.
type ProviderMetadata struct {
	ID           string
	Name         string
	Description  string
	EnvVar       string
	DefaultModel string
}

// GetProviderMetadata returns metadata for a given provider ID.
func GetProviderMetadata(id string) (ProviderMetadata, bool) {
	switch id {
	case "gemini":
		return ProviderMetadata{
			ID:           "gemini",
			Name:         "Google Gemini",
			Description:  "Fast, multimodal models from Google",
			EnvVar:       "GOOGLE_API_KEY",
			DefaultModel: "gemini-2.0-flash-exp",
		}, true
	case "openai":
		return ProviderMetadata{
			ID:           "openai",
			Name:         "OpenAI",
			Description:  "GPT-4o and other models from OpenAI",
			EnvVar:       "OPENAI_API_KEY",
			DefaultModel: "gpt-4o",
		}, true
	case "anthropic":
		return ProviderMetadata{
			ID:           "anthropic",
			Name:         "Anthropic Claude",
			Description:  "Claude 3.5 Sonnet and Haiku",
			EnvVar:       "ANTHROPIC_API_KEY",
			DefaultModel: "claude-3-5-sonnet-20241022",
		}, true
	case "ollama":
		return ProviderMetadata{
			ID:           "ollama",
			Name:         "Ollama",
			Description:  "Run LLMs locally",
			EnvVar:       "", // No API key required by default
			DefaultModel: "llama3",
		}, true
	default:
		return ProviderMetadata{}, false
	}
}

// GetSupportedProviders returns metadata for all supported providers.
func GetSupportedProviders() []ProviderMetadata {
	ids := provider.GetSupportedProviders()
	var metas []ProviderMetadata
	for _, id := range ids {
		if meta, ok := GetProviderMetadata(id); ok {
			metas = append(metas, meta)
		}
	}
	return metas
}

// GetAPIKey returns the API key for a provider from the environment.
func GetAPIKey(providerID string) string {
	if meta, ok := GetProviderMetadata(providerID); ok {
		return os.Getenv(meta.EnvVar)
	}
	return ""
}
