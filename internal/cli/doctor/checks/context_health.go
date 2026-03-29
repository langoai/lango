package checks

import (
	"context"
	"fmt"
	"math"
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
	// Validate context allocation ratios sum to 1.0 (±0.001, matching budget.go tolerance).
	alloc := cfg.Context.Allocation
	allocSum := alloc.Knowledge + alloc.RAG + alloc.Memory + alloc.RunSummary + alloc.Headroom
	if allocSum > 0 && math.Abs(allocSum-1.0) > 0.001 {
		warnings = append(warnings,
			fmt.Sprintf("context.allocation ratios sum to %.3f, should be 1.0 (±0.001)", allocSum))
	}
	if cfg.Embedding.RAG.Enabled && cfg.Embedding.Provider == "" {
		warnings = append(warnings,
			"embedding.rag.enabled=true but no embedding.provider configured")
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
