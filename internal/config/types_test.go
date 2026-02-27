package config

import "testing"

func TestResolveEmbeddingProvider_ByProviderMapKey(t *testing.T) {
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
			cfg := &Config{
				Embedding: EmbeddingConfig{Provider: tt.provider},
				Providers: tt.providers,
			}
			backend, apiKey := cfg.ResolveEmbeddingProvider()
			if backend != tt.wantBackend {
				t.Errorf("backend: want %q, got %q", tt.wantBackend, backend)
			}
			if (apiKey != "") != tt.wantHasAPIKey {
				t.Errorf("hasAPIKey: want %v, got apiKey=%q", tt.wantHasAPIKey, apiKey)
			}
		})
	}
}

func TestResolveEmbeddingProvider_LocalProvider(t *testing.T) {
	cfg := &Config{
		Embedding: EmbeddingConfig{Provider: "local"},
	}
	backend, apiKey := cfg.ResolveEmbeddingProvider()
	if backend != "local" {
		t.Errorf("backend: want %q, got %q", "local", backend)
	}
	if apiKey != "" {
		t.Errorf("apiKey: want empty, got %q", apiKey)
	}
}

func TestResolveEmbeddingProvider_NeitherConfigured(t *testing.T) {
	cfg := &Config{
		Embedding: EmbeddingConfig{},
	}
	backend, apiKey := cfg.ResolveEmbeddingProvider()
	if backend != "" {
		t.Errorf("backend: want empty, got %q", backend)
	}
	if apiKey != "" {
		t.Errorf("apiKey: want empty, got %q", apiKey)
	}
}

func TestResolveEmbeddingProvider_LegacyProviderIDFallback(t *testing.T) {
	// Legacy configs may still have ProviderID set. The resolver should
	// fall back to ProviderID when Provider is empty.
	cfg := &Config{
		Embedding: EmbeddingConfig{
			ProviderID: "gemini-1",
		},
		Providers: map[string]ProviderConfig{
			"gemini-1": {Type: "gemini", APIKey: "gemini-key"},
		},
	}

	backend, apiKey := cfg.ResolveEmbeddingProvider()
	if backend != "google" {
		t.Errorf("backend: want %q, got %q", "google", backend)
	}
	if apiKey != "gemini-key" {
		t.Errorf("apiKey: want %q, got %q", "gemini-key", apiKey)
	}
}

func TestMigrateEmbeddingProvider(t *testing.T) {
	t.Run("migrates ProviderID to Provider", func(t *testing.T) {
		cfg := &Config{
			Embedding: EmbeddingConfig{ProviderID: "my-openai"},
		}
		cfg.MigrateEmbeddingProvider()
		if cfg.Embedding.Provider != "my-openai" {
			t.Errorf("Provider: want %q, got %q", "my-openai", cfg.Embedding.Provider)
		}
		if cfg.Embedding.ProviderID != "" {
			t.Errorf("ProviderID should be empty after migration, got %q", cfg.Embedding.ProviderID)
		}
	})

	t.Run("Provider takes precedence when both set", func(t *testing.T) {
		cfg := &Config{
			Embedding: EmbeddingConfig{Provider: "local", ProviderID: "gemini-1"},
		}
		cfg.MigrateEmbeddingProvider()
		if cfg.Embedding.Provider != "local" {
			t.Errorf("Provider: want %q, got %q", "local", cfg.Embedding.Provider)
		}
		if cfg.Embedding.ProviderID != "" {
			t.Errorf("ProviderID should be empty after migration, got %q", cfg.Embedding.ProviderID)
		}
	})

	t.Run("no-op when only Provider is set", func(t *testing.T) {
		cfg := &Config{
			Embedding: EmbeddingConfig{Provider: "local"},
		}
		cfg.MigrateEmbeddingProvider()
		if cfg.Embedding.Provider != "local" {
			t.Errorf("Provider: want %q, got %q", "local", cfg.Embedding.Provider)
		}
	})
}
