package checks

import (
	"context"
	"testing"

	"github.com/langowarny/lango/internal/config"
)

func TestEmbeddingCheck_Run_ProviderIDResolvesCorrectly(t *testing.T) {
	cfg := &config.Config{
		Embedding: config.EmbeddingConfig{
			ProviderID: "gemini-1",
			Dimensions: 768,
		},
		Providers: map[string]config.ProviderConfig{
			"gemini-1": {Type: "gemini", APIKey: "test-key"},
		},
	}

	check := &EmbeddingCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass, got %v: %s", result.Status, result.Message)
	}
}

func TestEmbeddingCheck_Run_ProviderIDNotFound(t *testing.T) {
	cfg := &config.Config{
		Embedding: config.EmbeddingConfig{
			ProviderID: "nonexistent",
			Dimensions: 768,
		},
		Providers: map[string]config.ProviderConfig{
			"openai": {Type: "openai", APIKey: "sk-test"},
		},
	}

	check := &EmbeddingCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail, got %v: %s", result.Status, result.Message)
	}
}

func TestEmbeddingCheck_Run_ProviderIDNoAPIKey(t *testing.T) {
	cfg := &config.Config{
		Embedding: config.EmbeddingConfig{
			ProviderID: "my-openai",
			Dimensions: 1536,
		},
		Providers: map[string]config.ProviderConfig{
			"my-openai": {Type: "openai", APIKey: ""},
		},
	}

	check := &EmbeddingCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusFail {
		t.Errorf("expected StatusFail, got %v: %s", result.Status, result.Message)
	}
}

func TestEmbeddingCheck_Run_LocalProviderNoKey(t *testing.T) {
	cfg := &config.Config{
		Embedding: config.EmbeddingConfig{
			Provider:   "local",
			Dimensions: 768,
		},
	}

	check := &EmbeddingCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusPass {
		t.Errorf("expected StatusPass for local provider, got %v: %s", result.Status, result.Message)
	}
}

func TestEmbeddingCheck_Run_NeitherProviderConfigured(t *testing.T) {
	cfg := &config.Config{
		Embedding: config.EmbeddingConfig{},
	}

	check := &EmbeddingCheck{}
	result := check.Run(context.Background(), cfg)

	if result.Status != StatusSkip {
		t.Errorf("expected StatusSkip, got %v: %s", result.Status, result.Message)
	}
}
