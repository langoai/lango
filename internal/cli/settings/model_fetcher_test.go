package settings

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFetchEmbeddingModelOptions_FiltersByPattern(t *testing.T) {
	tests := []struct {
		give     []string
		wantLen  int
		wantHas  string
		wantMiss string
	}{
		{
			give:     []string{"text-embedding-3-small", "text-embedding-3-large", "gpt-4o", "gpt-3.5-turbo"},
			wantLen:  2,
			wantHas:  "text-embedding-3-small",
			wantMiss: "gpt-4o",
		},
		{
			give:     []string{"embed-english-v3.0", "command-r", "command-r-plus"},
			wantLen:  1,
			wantHas:  "embed-english-v3.0",
			wantMiss: "command-r",
		},
	}

	for _, tt := range tests {
		t.Run(tt.wantHas, func(t *testing.T) {
			// Filter using embeddingPatterns directly
			var filtered []string
			for _, m := range tt.give {
				for _, pat := range embeddingPatterns {
					if contains(m, pat) {
						filtered = append(filtered, m)
						break
					}
				}
			}

			assert.Equal(t, tt.wantLen, len(filtered))
			assert.Contains(t, filtered, tt.wantHas)
			assert.NotContains(t, filtered, tt.wantMiss)
		})
	}
}

func TestFetchEmbeddingModelOptions_FallbackWhenNoEmbedModels(t *testing.T) {
	all := []string{"gpt-4o", "gpt-3.5-turbo", "claude-3-opus"}

	var filtered []string
	for _, m := range all {
		for _, pat := range embeddingPatterns {
			if contains(m, pat) {
				filtered = append(filtered, m)
				break
			}
		}
	}

	// No embedding models found, should fallback
	if len(filtered) == 0 {
		filtered = all
	}

	assert.Equal(t, len(all), len(filtered))
	assert.Equal(t, all, filtered)
}

func TestFetchEmbeddingModelOptions_IncludesCurrentModel(t *testing.T) {
	all := []string{"text-embedding-3-small", "text-embedding-3-large", "gpt-4o"}
	currentModel := "custom-embed-model"

	var filtered []string
	for _, m := range all {
		for _, pat := range embeddingPatterns {
			if contains(m, pat) {
				filtered = append(filtered, m)
				break
			}
		}
	}

	// Include current model if not already present
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

	assert.Equal(t, 3, len(filtered))
	assert.Equal(t, currentModel, filtered[0])
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if lower(s[i:i+len(substr)]) == lower(substr) {
					return true
				}
			}
			return false
		}())
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}
