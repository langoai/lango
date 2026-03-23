package checks

import (
	"context"
	"fmt"
	"strings"

	"github.com/langoai/lango/internal/bootstrap"
	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/turntrace"
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

// RunWithBootstrap adds runtime diagnostics on top of static config validation.
func (c *MultiAgentCheck) RunWithBootstrap(
	ctx context.Context,
	cfg *config.Config,
	boot *bootstrap.Result,
) Result {
	base := c.Run(ctx, cfg)
	if base.Status == StatusSkip || base.Status == StatusFail {
		return base
	}
	if boot == nil || boot.DBClient == nil {
		base.Status = StatusWarn
		base.Details = "Runtime diagnostics unavailable: bootstrap DB client not available"
		return base
	}

	traceStore := turntrace.NewEntStore(boot.DBClient)
	failures, err := traceStore.RecentFailures(ctx, 3)
	if err != nil {
		base.Status = StatusWarn
		base.Details = fmt.Sprintf("Turn trace query failed: %v", err)
		return base
	}

	leakCount, err := traceStore.IsolationLeakCount(ctx, []string{
		"vault",
		"operator",
		"navigator",
		"librarian",
		"automator",
	})
	if err != nil {
		base.Status = StatusWarn
		base.Details = fmt.Sprintf("Isolation leak check failed: %v", err)
		return base
	}

	if len(failures) == 0 && leakCount == 0 {
		base.Message = fmt.Sprintf(
			"Multi-agent mode enabled (provider=%s, no recent failed traces, no isolation leaks)",
			cfg.Agent.Provider,
		)
		return base
	}

	base.Status = StatusWarn
	var details []string
	if len(failures) > 0 {
		details = append(details, "Recent failed traces:")
		for _, failure := range failures {
			details = append(details, fmt.Sprintf(
				"- %s [%s] %s",
				failure.TraceID,
				failure.Outcome,
				failure.Summary,
			))
		}
	}
	if leakCount > 0 {
		details = append(details, fmt.Sprintf(
			"Persisted raw isolated specialist turns detected: %d",
			leakCount,
		))
	}
	base.Message = fmt.Sprintf(
		"Multi-agent mode enabled with %d recent failed trace(s) and %d isolation leak row(s)",
		len(failures),
		leakCount,
	)
	base.Details = strings.Join(details, "\n")
	return base
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *MultiAgentCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
