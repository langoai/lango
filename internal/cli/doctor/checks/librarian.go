package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// LibrarianCheck validates the proactive librarian configuration.
type LibrarianCheck struct{}

// Name returns the check name.
func (c *LibrarianCheck) Name() string {
	return "Proactive Librarian"
}

// Run checks librarian configuration.
func (c *LibrarianCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Librarian.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Librarian is not enabled",
		}
	}

	var issues []string

	if !cfg.Knowledge.Enabled {
		issues = append(issues, "knowledge.enabled is false (librarian requires the knowledge store)")
	}

	if cfg.Librarian.Provider == "" && cfg.Agent.Provider == "" {
		issues = append(issues, "no provider configured for librarian analysis")
	}

	if len(issues) > 0 {
		message := "Librarian issues:\n"
		for _, issue := range issues {
			message += fmt.Sprintf("- %s\n", issue)
		}
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: message,
		}
	}

	provider := cfg.Librarian.Provider
	if provider == "" {
		provider = cfg.Agent.Provider
	}
	model := cfg.Librarian.Model
	if model == "" {
		model = "(default)"
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: fmt.Sprintf("Librarian enabled (provider=%s, model=%s)", provider, model),
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *LibrarianCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
