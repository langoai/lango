package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// EmbeddingCheck validates embedding/RAG configuration.
type EmbeddingCheck struct{}

// Name returns the check name.
func (c *EmbeddingCheck) Name() string {
	return "Embedding / RAG"
}

// Run checks embedding configuration validity.
func (c *EmbeddingCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	emb := cfg.Embedding
	if emb.Provider == "" && emb.ProviderID == "" {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Embedding provider not configured",
		}
	}

	var issues []string
	status := StatusPass

	// Resolve backend type and API key via unified resolver.
	backendType, apiKey := cfg.ResolveEmbeddingProvider()

	if backendType == "" {
		if emb.ProviderID != "" {
			issues = append(issues, fmt.Sprintf("provider ID %q not found in providers map or has unsupported type", emb.ProviderID))
		} else {
			issues = append(issues, "embedding provider not properly configured (set providerID or provider=local)")
		}
		status = StatusFail
	} else if backendType != "local" && apiKey == "" {
		issues = append(issues, fmt.Sprintf("provider %q has no API key configured", emb.ProviderID))
		status = StatusFail
	}

	// Check dimensions.
	if emb.Dimensions <= 0 {
		issues = append(issues, "dimensions should be set (e.g. 1536 for OpenAI, 768 for Google/local)")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Check RAG settings.
	if emb.RAG.Enabled && emb.RAG.MaxResults <= 0 {
		issues = append(issues, "rag.maxResults should be positive when RAG is enabled")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	if len(issues) == 0 {
		providerLabel := backendType
		if emb.ProviderID != "" {
			providerLabel = fmt.Sprintf("%s (id=%s)", backendType, emb.ProviderID)
		}
		msg := fmt.Sprintf("Embedding configured (provider=%s, dimensions=%d, rag=%v)",
			providerLabel, emb.Dimensions, emb.RAG.Enabled)
		return Result{Name: c.Name(), Status: StatusPass, Message: msg}
	}

	message := "Embedding issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *EmbeddingCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
