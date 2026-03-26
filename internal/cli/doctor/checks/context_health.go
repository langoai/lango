package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// ContextHealthCheck validates the overall context engineering configuration.
type ContextHealthCheck struct{}

// Name returns the check name.
func (c *ContextHealthCheck) Name() string {
	return "Context Engineering"
}

// Run checks context subsystem configuration and consistency.
func (c *ContextHealthCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	var warnings []string

	// Check if a context profile is set.
	if cfg.ContextProfile == "" {
		warnings = append(warnings,
			"no contextProfile set — consider using 'balanced' or 'full' for managed defaults")
	}

	// Count enabled context subsystems.
	type subsystem struct {
		name    string
		enabled bool
	}
	subs := []subsystem{
		{"Knowledge", cfg.Knowledge.Enabled},
		{"Obs. Memory", cfg.ObservationalMemory.Enabled},
		{"Graph", cfg.Graph.Enabled},
		{"Librarian", cfg.Librarian.Enabled},
	}

	enabledCount := 0
	for _, s := range subs {
		if s.enabled {
			enabledCount++
		}
	}

	// Check embedding provider.
	hasEmbedding := cfg.Embedding.Provider != ""
	if !hasEmbedding && enabledCount > 0 {
		warnings = append(warnings,
			"embedding.provider is not configured — knowledge and RAG features will not work")
	}

	// Detect silently disabled features (dependencies missing).
	if cfg.Librarian.Enabled && !cfg.Knowledge.Enabled {
		warnings = append(warnings,
			"librarian is enabled but knowledge is disabled — librarian cannot store results")
	}
	if cfg.Graph.Enabled && !hasEmbedding {
		warnings = append(warnings,
			"graph is enabled but embedding is not configured — graph RAG will be unavailable")
	}

	if len(warnings) == 0 {
		profile := string(cfg.ContextProfile)
		if profile == "" {
			profile = "(none)"
		}
		return Result{
			Name:   c.Name(),
			Status: StatusPass,
			Message: fmt.Sprintf("Context OK — profile=%s, %d/%d subsystems enabled, embedding=%s",
				profile, enabledCount, len(subs), cfg.Embedding.Provider),
		}
	}

	profile := string(cfg.ContextProfile)
	if profile == "" {
		profile = "(none)"
	}
	details := strings.Join(warnings, "\n")
	return Result{
		Name:   c.Name(),
		Status: StatusWarn,
		Message: fmt.Sprintf("Context issues — profile=%s, %d/%d subsystems enabled",
			profile, enabledCount, len(subs)),
		Details: details,
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ContextHealthCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
