package provider

import (
	"errors"
	"fmt"
	"strings"
)

// ErrModelProviderMismatch indicates a model name belongs to a different provider.
var ErrModelProviderMismatch = errors.New("model-provider mismatch")

// modelExclusions maps provider types to model prefixes that clearly belong
// to a *different* provider. This is a heuristic safety net, not a complete
// model registry — it catches obvious cross-provider routing errors like
// sending "gpt-5.3-codex" to the Gemini API.
var modelExclusions = map[string][]string{
	"openai":    {"claude-", "gemini-", "models/gemini-"},
	"anthropic": {"gpt-", "o1-", "o3-", "o4-", "chatgpt-", "gemini-", "models/gemini-"},
	"gemini":    {"gpt-", "o1-", "o3-", "o4-", "chatgpt-", "claude-"},
	"google":    {"gpt-", "o1-", "o3-", "o4-", "chatgpt-", "claude-"},
	// ollama and github can host any model — no exclusions.
}

// ValidateModelProvider checks whether the given model name is clearly
// incompatible with the provider type. Returns ErrModelProviderMismatch
// when the model prefix matches a known exclusion for the provider.
//
// An empty model is always valid (the provider will use its default).
func ValidateModelProvider(providerType, model string) error {
	if model == "" {
		return nil
	}

	lowerModel := strings.ToLower(model)
	exclusions, ok := modelExclusions[strings.ToLower(providerType)]
	if !ok {
		return nil
	}

	for _, prefix := range exclusions {
		if strings.HasPrefix(lowerModel, prefix) {
			return fmt.Errorf("%w: model %q is not compatible with provider type %q", ErrModelProviderMismatch, model, providerType)
		}
	}
	return nil
}
