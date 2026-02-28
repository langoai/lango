package settings

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/provider"
	provanthropic "github.com/langoai/lango/internal/provider/anthropic"
	provgemini "github.com/langoai/lango/internal/provider/gemini"
	provopenai "github.com/langoai/lango/internal/provider/openai"
	"github.com/langoai/lango/internal/types"
)

const modelFetchTimeout = 15 * time.Second

// NewProviderFromConfig creates a lightweight provider instance from config.
// Returns nil if the provider cannot be created (missing API key, unknown type, etc.).
func NewProviderFromConfig(id string, pCfg config.ProviderConfig) provider.Provider {
	apiKey := pCfg.APIKey
	if apiKey == "" && pCfg.Type != types.ProviderOllama {
		return nil
	}

	switch pCfg.Type {
	case types.ProviderOpenAI:
		return provopenai.NewProvider(id, apiKey, pCfg.BaseURL)
	case types.ProviderAnthropic:
		return provanthropic.NewProvider(id, apiKey)
	case types.ProviderGemini, types.ProviderGoogle:
		p, err := provgemini.NewProvider(context.Background(), id, apiKey, "")
		if err != nil {
			return nil
		}
		return p
	case types.ProviderOllama:
		baseURL := pCfg.BaseURL
		if baseURL == "" {
			baseURL = "http://localhost:11434/v1"
		}
		return provopenai.NewProvider(id, apiKey, baseURL)
	case types.ProviderGitHub:
		baseURL := pCfg.BaseURL
		if baseURL == "" {
			baseURL = "https://models.inference.ai.azure.com"
		}
		return provopenai.NewProvider(id, apiKey, baseURL)
	default:
		return nil
	}
}

// FetchModelOptions fetches available models from a provider.
// Returns a sorted list of model IDs, or nil if fetching fails.
// The currentModel (if non-empty) is always included in the result.
func FetchModelOptions(providerID string, cfg *config.Config, currentModel string) []string {
	opts, _ := FetchModelOptionsWithError(providerID, cfg, currentModel)
	return opts
}

// FetchModelOptionsWithError is like FetchModelOptions but also returns
// the error when model fetching fails, for diagnostic feedback.
func FetchModelOptionsWithError(providerID string, cfg *config.Config, currentModel string) ([]string, error) {
	pCfg, ok := cfg.Providers[providerID]
	if !ok {
		return nil, fmt.Errorf("provider %q not found in config", providerID)
	}

	p := NewProviderFromConfig(providerID, pCfg)
	if p == nil {
		return nil, fmt.Errorf("provider %q: missing API key or unsupported type", providerID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), modelFetchTimeout)
	defer cancel()

	models, err := p.ListModels(ctx)
	if err != nil {
		return nil, fmt.Errorf("provider %q: %w", providerID, err)
	}
	if len(models) == 0 {
		return nil, fmt.Errorf("provider %q returned no models", providerID)
	}

	seen := make(map[string]bool, len(models))
	opts := make([]string, 0, len(models))
	for _, m := range models {
		if !seen[m.ID] {
			seen[m.ID] = true
			opts = append(opts, m.ID)
		}
	}
	sort.Strings(opts)

	// Ensure current model is included
	if currentModel != "" && !seen[currentModel] {
		opts = append([]string{currentModel}, opts...)
	}

	return opts, nil
}

// embeddingPatterns contains substrings that indicate embedding models.
var embeddingPatterns = []string{"embed", "embedding"}

// FetchEmbeddingModelOptions fetches models and filters for embedding-capable ones.
// Falls back to the full model list if no embedding models are found.
func FetchEmbeddingModelOptions(providerID string, cfg *config.Config, currentModel string) []string {
	all := FetchModelOptions(providerID, cfg, currentModel)
	if len(all) == 0 {
		return nil
	}

	var filtered []string
	for _, m := range all {
		lower := strings.ToLower(m)
		for _, pat := range embeddingPatterns {
			if strings.Contains(lower, pat) {
				filtered = append(filtered, m)
				break
			}
		}
	}

	// Ensure current model is included in filtered results
	if currentModel != "" && len(filtered) > 0 {
		found := false
		for _, m := range filtered {
			if m == currentModel {
				found = true
				break
			}
		}
		if !found {
			filtered = append([]string{currentModel}, filtered...)
		}
	}

	// Fallback to full list if no embedding models detected
	if len(filtered) == 0 {
		return all
	}
	return filtered
}
