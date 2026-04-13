package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/langoai/lango/internal/config"
)

func TestContextHealthCheck_Run(t *testing.T) {
	tests := []struct {
		give       string
		cfg        *config.Config
		wantStatus Status
		wantSubstr string
	}{
		{
			give: "balanced profile all subsystems OK",
			cfg: &config.Config{
				ContextProfile:      config.ContextProfileBalanced,
				Knowledge:           config.KnowledgeConfig{Enabled: true},
				ObservationalMemory: config.ObservationalMemoryConfig{Enabled: true},
				Librarian:           config.LibrarianConfig{Enabled: true},
				Embedding:           config.EmbeddingConfig{Provider: "openai"},
			},
			wantStatus: StatusPass,
			wantSubstr: "balanced",
		},
		{
			give: "full profile everything enabled",
			cfg: &config.Config{
				ContextProfile:      config.ContextProfileFull,
				Knowledge:           config.KnowledgeConfig{Enabled: true},
				ObservationalMemory: config.ObservationalMemoryConfig{Enabled: true},
				Librarian:           config.LibrarianConfig{Enabled: true},
				Graph:               config.GraphConfig{Enabled: true},
				Embedding:           config.EmbeddingConfig{Provider: "gemini"},
			},
			wantStatus: StatusPass,
			wantSubstr: "4/4 subsystems enabled",
		},
		{
			give:       "no profile nothing enabled",
			cfg:        &config.Config{},
			wantStatus: StatusWarn,
			wantSubstr: "no contextProfile set",
		},
		{
			give: "profile set but embedding missing",
			cfg: &config.Config{
				ContextProfile:      config.ContextProfileBalanced,
				Knowledge:           config.KnowledgeConfig{Enabled: true},
				ObservationalMemory: config.ObservationalMemoryConfig{Enabled: true},
				Librarian:           config.LibrarianConfig{Enabled: true},
			},
			wantStatus: StatusWarn,
			wantSubstr: "embedding.provider is not configured",
		},
		{
			give: "librarian enabled but knowledge disabled",
			cfg: &config.Config{
				ContextProfile: config.ContextProfileBalanced,
				Librarian:      config.LibrarianConfig{Enabled: true},
				Embedding:      config.EmbeddingConfig{Provider: "openai"},
			},
			wantStatus: StatusWarn,
			wantSubstr: "librarian is enabled but knowledge is disabled",
		},
		{
			give: "graph enabled but no embedding",
			cfg: &config.Config{
				ContextProfile: config.ContextProfileFull,
				Knowledge:      config.KnowledgeConfig{Enabled: true},
				Graph:          config.GraphConfig{Enabled: true},
			},
			wantStatus: StatusWarn,
			wantSubstr: "graph is enabled but embedding is not configured",
		},
		{
			give:       "nil config",
			cfg:        nil,
			wantStatus: StatusSkip,
			wantSubstr: "not loaded",
		},
	}

	check := &ContextHealthCheck{}
	assert.Equal(t, "Context Engineering", check.Name())

	for _, tt := range tests {
		t.Run(tt.give, func(t *testing.T) {
			result := check.Run(context.Background(), tt.cfg)
			assert.Equal(t, tt.wantStatus, result.Status, "status mismatch: %s", result.Message)
			combined := result.Message + " " + result.Details
			assert.Contains(t, combined, tt.wantSubstr)
		})
	}
}

func TestContextHealthCheck_Fix_DelegatesToRun(t *testing.T) {
	check := &ContextHealthCheck{}
	cfg := &config.Config{}

	run := check.Run(context.Background(), cfg)
	fix := check.Fix(context.Background(), cfg)
	assert.Equal(t, run.Status, fix.Status)
	assert.Equal(t, run.Message, fix.Message)
}
