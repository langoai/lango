package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// GraphStoreCheck validates graph store configuration.
type GraphStoreCheck struct{}

// Name returns the check name.
func (c *GraphStoreCheck) Name() string {
	return "Graph Store"
}

// Run checks graph store configuration validity.
func (c *GraphStoreCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Graph.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Graph store is not enabled",
		}
	}

	var issues []string
	status := StatusPass

	if cfg.Graph.Backend != "bolt" {
		issues = append(issues, fmt.Sprintf("unsupported backend %q (must be \"bolt\")", cfg.Graph.Backend))
		status = StatusFail
	}

	if cfg.Graph.DatabasePath == "" {
		issues = append(issues, "graph.databasePath is not set")
		status = StatusFail
	}

	if cfg.Graph.MaxTraversalDepth <= 0 {
		issues = append(issues, "graph.maxTraversalDepth should be positive")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	if cfg.Graph.MaxExpansionResults <= 0 {
		issues = append(issues, "graph.maxExpansionResults should be positive")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	if len(issues) == 0 {
		msg := fmt.Sprintf("Graph store configured (backend=%s, path=%s, depth=%d, expand=%d)",
			cfg.Graph.Backend, cfg.Graph.DatabasePath,
			cfg.Graph.MaxTraversalDepth, cfg.Graph.MaxExpansionResults)
		return Result{Name: c.Name(), Status: StatusPass, Message: msg}
	}

	message := "Graph store issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *GraphStoreCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
