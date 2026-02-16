package config

import "testing"

func TestResolveEmbeddingProvider_ExplicitProviderID(t *testing.T) {
	tests := []struct {
		give           string
		providerID     string
		providers      map[string]ProviderConfig
		wantBackend    string
		wantHasAPIKey  bool
	}{
		{
			give:       "gemini provider by custom ID",
			providerID: "gemini-1",
			providers: map[string]ProviderConfig{
				"gemini-1": {Type: "gemini", APIKey: "test-key"},
			},
			wantBackend:   "google",
			wantHasAPIKey: true,
		},
		{
			give:       "openai provider by custom ID",
			providerID: "my-openai",
			providers: map[string]ProviderConfig{
				"my-openai": {Type: "openai", APIKey: "sk-test"},
			},
			wantBackend:   "openai",
			wantHasAPIKey: true,
		},
		{
			give:       "ollama provider by custom ID",
			providerID: "my-ollama",
			providers: map[string]ProviderConfig{
				"my-ollama": {Type: "ollama"},
			},
			wantBackend:   "local",
			wantHasAPIKey: false,
		},
		{
			give:       "anthropic provider has no embedding support",
			providerID: "my-claude",
			providers: map[string]ProviderConfig{
				"my-claude": {Type: "anthropic", APIKey: "sk-ant-test"},
			},
			wantBackend:   "",
			wantHasAPIKey: false,
		},
		{
			give:       "provider ID not found",
			providerID: "nonexistent",
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
				Embedding: EmbeddingConfig{ProviderID: tt.providerID},
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

func TestResolveEmbeddingProvider_ProviderIDTakesPrecedence(t *testing.T) {
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
