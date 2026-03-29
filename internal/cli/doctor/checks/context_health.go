package checks

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/langoai/lango/internal/config"
)

// ContextHealthCheck validates context engineering configuration.
type ContextHealthCheck struct{}

// Name returns the check name.
func (c *ContextHealthCheck) Name() string {
	return "Context Health"
}

// Run checks context engineering configuration validity.
func (c *ContextHealthCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Knowledge.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Knowledge system disabled — context health checks skipped",
		}
	}

	var warnings []string

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

	if len(warnings) > 0 {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: fmt.Sprintf("Context health: %d warning(s)", len(warnings)),
			Details: strings.Join(warnings, "; "),
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Context engineering configuration valid",
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ContextHealthCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
