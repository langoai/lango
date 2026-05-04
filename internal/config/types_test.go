package config

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigClone_DeepCopy(t *testing.T) {
	orig := DefaultConfig()
	orig.Agent.Provider = "openai"
	orig.Agent.Model = "gpt-4"

	clone := orig.Clone()
	require.NotNil(t, clone)

	// Values are equal.
	assert.Equal(t, "openai", clone.Agent.Provider)
	assert.Equal(t, "gpt-4", clone.Agent.Model)

	// Mutation of clone does not affect original.
	clone.Agent.Provider = "anthropic"
	assert.Equal(t, "openai", orig.Agent.Provider)
}

func TestConfigClone_NilSafe(t *testing.T) {
	var c *Config
	assert.Nil(t, c.Clone())
}

func TestConfigClone_PreservesSparseOntologyGovernanceAbsence(t *testing.T) {
	t.Parallel()

	orig := &Config{
		Ontology: OntologyConfig{
			Governance: OntologyGovernanceConfig{},
		},
	}

	clone := orig.Clone()
	require.NotNil(t, clone)
	assert.Empty(t, clone.Ontology.Governance.AdmissionMode)
	assert.Equal(t, 0.0, clone.Ontology.Governance.LearningDefaultConfidence)
	assert.Equal(t, 0.0, clone.Ontology.Governance.LibrarianDefaultConfidence)
	assert.False(t, clone.Ontology.Governance.LearningDefaultConfidenceBackfillNeeded)
	assert.False(t, clone.Ontology.Governance.LibrarianDefaultConfidenceBackfillNeeded)
	assert.False(t, clone.Ontology.Governance.LearningDefaultConfidencePresent)
	assert.False(t, clone.Ontology.Governance.LibrarianDefaultConfidencePresent)
}

func TestConfigClone_PreservesSparseAdmissionModeOffSemantics(t *testing.T) {
	t.Parallel()

	orig := &Config{
		Ontology: OntologyConfig{
			Governance: OntologyGovernanceConfig{
				AdmissionMode: OntologyAdmissionModeOff,
			},
		},
	}

	clone := orig.Clone()
	require.NotNil(t, clone)
	assert.Equal(t, OntologyAdmissionModeOff, clone.Ontology.Governance.AdmissionMode)
	assert.False(t, clone.Ontology.Governance.AdmissionModePresent)
}

func TestConfigClone_PreservesExplicitZeroAdmissionConfidenceMarkers(t *testing.T) {
	t.Parallel()

	orig := &Config{
		Ontology: OntologyConfig{
			Governance: OntologyGovernanceConfig{
				AdmissionMode:                            OntologyAdmissionModeOff,
				LearningDefaultConfidence:                0.0,
				LearningDefaultConfidenceBackfillNeeded:  true,
				LearningDefaultConfidencePresent:         true,
				LibrarianDefaultConfidence:               0.0,
				LibrarianDefaultConfidenceBackfillNeeded: true,
				LibrarianDefaultConfidencePresent:        true,
			},
		},
	}

	clone := orig.Clone()
	require.NotNil(t, clone)
	assert.Equal(t, OntologyAdmissionModeOff, clone.Ontology.Governance.AdmissionMode)
	assert.Equal(t, 0.0, clone.Ontology.Governance.LearningDefaultConfidence)
	assert.Equal(t, 0.0, clone.Ontology.Governance.LibrarianDefaultConfidence)
	assert.True(t, clone.Ontology.Governance.LearningDefaultConfidenceBackfillNeeded)
	assert.True(t, clone.Ontology.Governance.LibrarianDefaultConfidenceBackfillNeeded)
	assert.True(t, clone.Ontology.Governance.LearningDefaultConfidencePresent)
	assert.True(t, clone.Ontology.Governance.LibrarianDefaultConfidencePresent)
}

func TestOntologyGovernanceJSONMarshal_OmitsSparseConfidenceKeys(t *testing.T) {
	t.Parallel()

	var cfg Config
	require.NoError(t, json.Unmarshal([]byte(`{"ontology":{"governance":{"admissionMode":"off"}}}`), &cfg))

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"admissionMode":"off"`)
	assert.NotContains(t, string(data), `learningDefaultConfidence`)
	assert.NotContains(t, string(data), `librarianDefaultConfidence`)
}

func TestOntologyGovernanceJSONMarshal_OmitsSparseAdmissionModeKey(t *testing.T) {
	t.Parallel()

	var cfg Config
	require.NoError(t, json.Unmarshal([]byte(`{"ontology":{"governance":{}}}`), &cfg))

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.NotContains(t, string(data), `"admissionMode"`)
}

func TestOntologyGovernanceJSONMarshal_PreservesExplicitOffAdmissionMode(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Ontology: OntologyConfig{
			Governance: OntologyGovernanceConfig{
				AdmissionMode:        OntologyAdmissionModeOff,
				AdmissionModePresent: true,
			},
		},
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"admissionMode":"off"`)

	var decoded Config
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, OntologyAdmissionModeOff, decoded.Ontology.Governance.AdmissionMode)
	assert.True(t, decoded.Ontology.Governance.AdmissionModePresent)
}

func TestOntologyGovernanceJSONMarshal_PreservesExplicitZeroConfidenceKeys(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Ontology: OntologyConfig{
			Governance: OntologyGovernanceConfig{
				AdmissionMode:                     OntologyAdmissionModeOff,
				LearningDefaultConfidence:         0.0,
				LearningDefaultConfidencePresent:  true,
				LibrarianDefaultConfidence:        0.0,
				LibrarianDefaultConfidencePresent: true,
			},
		},
	}

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"learningDefaultConfidence":0`)
	assert.Contains(t, string(data), `"librarianDefaultConfidence":0`)
}

func TestOntologyGovernanceJSONMarshal_OmitsZeroValueSparseAdmissionKeys(t *testing.T) {
	t.Parallel()

	data, err := json.Marshal(OntologyGovernanceConfig{})
	require.NoError(t, err)
	assert.NotContains(t, string(data), `admissionMode`)
	assert.NotContains(t, string(data), `learningDefaultConfidence`)
	assert.NotContains(t, string(data), `librarianDefaultConfidence`)
}

func TestOntologyGovernanceJSONMarshal_PreservesDefaultConfigZeroConfidenceRoundTrip(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()
	cfg.Ontology.Governance.LearningDefaultConfidence = 0.0
	cfg.Ontology.Governance.LibrarianDefaultConfidence = 0.0

	data, err := json.Marshal(cfg)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"learningDefaultConfidence":0`)
	assert.Contains(t, string(data), `"librarianDefaultConfidence":0`)

	var decoded Config
	require.NoError(t, json.Unmarshal(data, &decoded))
	assert.Equal(t, 0.0, decoded.Ontology.Governance.LearningDefaultConfidence)
	assert.Equal(t, 0.0, decoded.Ontology.Governance.LibrarianDefaultConfidence)
}

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
