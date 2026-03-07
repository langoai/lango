package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolveEmbeddingProvider_ByProviderMapKey(t *testing.T) {
	t.Parallel()

	tests := []struct {
		give          string
		provider      string
		providers     map[string]ProviderConfig
		wantBackend   string
		wantHasAPIKey bool
	}{
		{
			give:     "gemini provider by custom ID",
			provider: "gemini-1",
			providers: map[string]ProviderConfig{
				"gemini-1": {Type: "gemini", APIKey: "test-key"},
			},
			wantBackend:   "google",
			wantHasAPIKey: true,
		},
		{
			give:     "openai provider by custom ID",
			provider: "my-openai",
			providers: map[string]ProviderConfig{
				"my-openai": {Type: "openai", APIKey: "sk-test"},
			},
			wantBackend:   "openai",
			wantHasAPIKey: true,
		},
		{
			give:     "ollama provider by custom ID",
			provider: "my-ollama",
			providers: map[string]ProviderConfig{
				"my-ollama": {Type: "ollama"},
			},
			wantBackend:   "local",
			wantHasAPIKey: false,
		},
		{
			give:     "anthropic provider has no embedding support",
			provider: "my-claude",
			providers: map[string]ProviderConfig{
				"my-claude": {Type: "anthropic", APIKey: "sk-ant-test"},
			},
			wantBackend:   "",
			wantHasAPIKey: false,
		},
		{
			give:     "provider not found",
			provider: "nonexistent",
			providers: map[string]ProviderConfig{
				"openai": {Type: "openai", APIKey: "sk-test"},
			},
			wantBackend:   "",
			wantHasAPIKey: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			t.Parallel()

			cfg := &Config{
				Embedding: EmbeddingConfig{Provider: tt.provider},
				Providers: tt.providers,
			}
			backend, apiKey := cfg.ResolveEmbeddingProvider()
			assert.Equal(t, tt.wantBackend, backend, "backend")
			assert.Equal(t, tt.wantHasAPIKey, apiKey != "", "hasAPIKey")
		})
	}
}

func TestResolveEmbeddingProvider_LocalProvider(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Embedding: EmbeddingConfig{Provider: "local"},
	}
	backend, apiKey := cfg.ResolveEmbeddingProvider()
	assert.Equal(t, "local", backend)
	assert.Empty(t, apiKey)
}

func TestResolveEmbeddingProvider_NeitherConfigured(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Embedding: EmbeddingConfig{},
	}
	backend, apiKey := cfg.ResolveEmbeddingProvider()
	assert.Empty(t, backend)
	assert.Empty(t, apiKey)
}

func TestResolveEmbeddingProvider_LegacyProviderIDFallback(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Embedding: EmbeddingConfig{
			ProviderID: "gemini-1",
		},
		Providers: map[string]ProviderConfig{
			"gemini-1": {Type: "gemini", APIKey: "gemini-key"},
		},
	}

	backend, apiKey := cfg.ResolveEmbeddingProvider()
	assert.Equal(t, "google", backend)
	assert.Equal(t, "gemini-key", apiKey)
}

func TestMigrateEmbeddingProvider(t *testing.T) {
	t.Parallel()

	t.Run("migrates ProviderID to Provider", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Embedding: EmbeddingConfig{ProviderID: "my-openai"},
		}
		cfg.MigrateEmbeddingProvider()
		assert.Equal(t, "my-openai", cfg.Embedding.Provider)
		assert.Empty(t, cfg.Embedding.ProviderID)
	})

	t.Run("Provider takes precedence when both set", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Embedding: EmbeddingConfig{Provider: "local", ProviderID: "gemini-1"},
		}
		cfg.MigrateEmbeddingProvider()
		assert.Equal(t, "local", cfg.Embedding.Provider)
		assert.Empty(t, cfg.Embedding.ProviderID)
	})

	t.Run("no-op when only Provider is set", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Embedding: EmbeddingConfig{Provider: "local"},
		}
		cfg.MigrateEmbeddingProvider()
		assert.Equal(t, "local", cfg.Embedding.Provider)
	})

	t.Run("migrates Local.Model to Model", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Embedding: EmbeddingConfig{
				Provider: "local",
				Local:    LocalEmbeddingConfig{Model: "nomic-embed-text"},
			},
		}
		cfg.MigrateEmbeddingProvider()
		assert.Equal(t, "nomic-embed-text", cfg.Embedding.Model)
		assert.Empty(t, cfg.Embedding.Local.Model)
	})

	t.Run("Model takes precedence over Local.Model", func(t *testing.T) {
		t.Parallel()

		cfg := &Config{
			Embedding: EmbeddingConfig{
				Provider: "local",
				Model:    "text-embedding-3-small",
				Local:    LocalEmbeddingConfig{Model: "nomic-embed-text"},
			},
		}
		cfg.MigrateEmbeddingProvider()
		assert.Equal(t, "text-embedding-3-small", cfg.Embedding.Model)
		assert.Empty(t, cfg.Embedding.Local.Model)
	})
}
