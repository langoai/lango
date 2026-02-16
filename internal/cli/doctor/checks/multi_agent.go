package checks

import (
	"context"
	"fmt"

	"github.com/langowarny/lango/internal/config"
)

// MultiAgentCheck validates multi-agent orchestration configuration.
type MultiAgentCheck struct{}

// Name returns the check name.
func (c *MultiAgentCheck) Name() string {
	return "Multi-Agent"
}

// Run checks multi-agent configuration validity.
func (c *MultiAgentCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Agent.MultiAgent {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Multi-agent mode is not enabled",
		}
	}

	var issues []string
	status := StatusPass

	if cfg.Agent.Provider == "" {
		issues = append(issues, "agent.provider is required for multi-agent mode")
		status = StatusFail
	}

	if len(issues) == 0 {
		msg := fmt.Sprintf("Multi-agent mode enabled (provider=%s)", cfg.Agent.Provider)
		return Result{Name: c.Name(), Status: StatusPass, Message: msg}
	}

	message := "Multi-agent issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *MultiAgentCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
