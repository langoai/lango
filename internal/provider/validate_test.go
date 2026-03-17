package provider

import (
	"errors"
	"testing"
)

func TestValidateModelProvider(t *testing.T) {
	tests := []struct {
		give         string
		providerType string
		model        string
		wantErr      bool
	}{
		// Empty model is always valid.
		{give: "empty model", providerType: "gemini", model: "", wantErr: false},

		// Gemini provider with correct models.
		{give: "gemini+gemini-3-flash", providerType: "gemini", model: "gemini-3-flash-preview", wantErr: false},
		{give: "gemini+models/gemini", providerType: "gemini", model: "gemini-2.0-flash", wantErr: false},

		// Gemini provider with wrong models.
		{give: "gemini+gpt", providerType: "gemini", model: "gpt-5.3-codex", wantErr: true},
		{give: "gemini+claude", providerType: "gemini", model: "claude-sonnet-4-5-20250514", wantErr: true},
		{give: "gemini+o4", providerType: "gemini", model: "o4-mini", wantErr: true},

		// OpenAI provider with correct models.
		{give: "openai+gpt", providerType: "openai", model: "gpt-5.3-codex", wantErr: false},
		{give: "openai+o4", providerType: "openai", model: "o4-mini", wantErr: false},

		// OpenAI provider with wrong models.
		{give: "openai+gemini", providerType: "openai", model: "gemini-3-flash-preview", wantErr: true},
		{give: "openai+claude", providerType: "openai", model: "claude-sonnet-4-5-20250514", wantErr: true},
		{give: "openai+models/gemini", providerType: "openai", model: "models/gemini-2.0-flash", wantErr: true},

		// Anthropic provider with correct models.
		{give: "anthropic+claude", providerType: "anthropic", model: "claude-sonnet-4-5-20250514", wantErr: false},

		// Anthropic provider with wrong models.
		{give: "anthropic+gpt", providerType: "anthropic", model: "gpt-5.3-codex", wantErr: true},
		{give: "anthropic+gemini", providerType: "anthropic", model: "gemini-3-flash-preview", wantErr: true},

		// Google alias works the same as gemini.
		{give: "google+gemini", providerType: "google", model: "gemini-3-flash-preview", wantErr: false},
		{give: "google+gpt", providerType: "google", model: "gpt-4o", wantErr: true},

		// Ollama can host any model.
		{give: "ollama+gpt", providerType: "ollama", model: "gpt-4o", wantErr: false},
		{give: "ollama+claude", providerType: "ollama", model: "claude-sonnet-4-5-20250514", wantErr: false},
		{give: "ollama+gemini", providerType: "ollama", model: "gemini-3-flash-preview", wantErr: false},

		// GitHub can host any model.
		{give: "github+gpt", providerType: "github", model: "gpt-4o", wantErr: false},

		// Unknown provider — no exclusions.
		{give: "unknown+anything", providerType: "custom", model: "my-model", wantErr: false},

		// Case insensitivity.
		{give: "gemini+GPT-upper", providerType: "gemini", model: "GPT-5.3-codex", wantErr: true},
		{give: "OPENAI+claude", providerType: "OPENAI", model: "claude-sonnet-4-5-20250514", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			err := ValidateModelProvider(tt.providerType, tt.model)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for provider=%q model=%q, got nil", tt.providerType, tt.model)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error for provider=%q model=%q: %v", tt.providerType, tt.model, err)
			}
			if tt.wantErr && err != nil && !errors.Is(err, ErrModelProviderMismatch) {
				t.Errorf("expected ErrModelProviderMismatch, got: %v", err)
			}
		})
	}
}
