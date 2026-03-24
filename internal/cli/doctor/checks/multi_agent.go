package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

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

	// Extended diagnostics: loop/timeout frequency, trace growth, avg duration.
	extWarnings := runExtendedTraceChecks(ctx, traceStore, cfg)

	if len(failures) == 0 && leakCount == 0 && len(extWarnings) == 0 {
		base.Message = fmt.Sprintf(
			"Multi-agent mode enabled (provider=%s, no recent failed traces, no isolation leaks)",
			cfg.Agent.Provider,
		)
		return base
	}

	base.Status = StatusWarn
	var details []string

	for _, w := range extWarnings {
		details = append(details, w)
	}
	base.TraceFailures = make([]TraceFailure, 0, len(failures))
	if len(failures) > 0 {
		details = append(details, "Recent failed traces:")
		for _, failure := range failures {
			base.TraceFailures = append(base.TraceFailures, TraceFailure{
				TraceID:    failure.TraceID,
				Outcome:    string(failure.Outcome),
				ErrorCode:  failure.ErrorCode,
				CauseClass: failure.CauseClass,
				Summary:    failure.Summary,
			})
			details = append(details, fmt.Sprintf(
				"- %s [%s/%s/%s] %s",
				failure.TraceID,
				failure.Outcome,
				failure.ErrorCode,
				failure.CauseClass,
				failure.Summary,
			))
		}
	}
	if leakCount > 0 {
		base.IsolationLeakCount = &leakCount
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

// runExtendedTraceChecks performs additional trace-based diagnostics.
func runExtendedTraceChecks(ctx context.Context, store *turntrace.EntStore, cfg *config.Config) []string {
	var warnings []string
	since := time.Now().Add(-24 * time.Hour)

	// Loop frequency: warn if >3 loop_detected in 24h.
	loops, err := store.RecentByOutcome(ctx, turntrace.OutcomeLoopDetected, since, 100)
	if err == nil && len(loops) > 3 {
		warnings = append(warnings, fmt.Sprintf(
			"High loop frequency: %d loop_detected traces in last 24h (threshold: 3)",
			len(loops),
		))
	}

	// Timeout frequency: warn if >5 timeout in 24h.
	timeouts, err := store.RecentByOutcome(ctx, turntrace.OutcomeTimeout, since, 100)
	if err == nil && len(timeouts) > 5 {
		warnings = append(warnings, fmt.Sprintf(
			"High timeout frequency: %d timeout traces in last 24h (threshold: 5)",
			len(timeouts),
		))
	}

	// Trace store growth: warn if >80% of maxTraces.
	maxTraces := cfg.Observability.TraceStore.MaxTraces
	if maxTraces <= 0 {
		maxTraces = 10000
	}
	count, err := store.TraceCount(ctx)
	if err == nil && count > maxTraces*80/100 {
		warnings = append(warnings, fmt.Sprintf(
			"Trace store at %d%% capacity (%d/%d)",
			count*100/maxTraces, count, maxTraces,
		))
	}

	// Average turn duration: warn if >2min.
	successes, err := store.RecentByOutcome(ctx, turntrace.OutcomeSuccess, since, 100)
	if err == nil && len(successes) > 0 {
		var totalDur time.Duration
		counted := 0
		for _, t := range successes {
			if t.EndedAt != nil {
				totalDur += t.EndedAt.Sub(t.StartedAt)
				counted++
			}
		}
		if counted > 0 {
			avg := totalDur / time.Duration(counted)
			if avg > 2*time.Minute {
				warnings = append(warnings, fmt.Sprintf(
					"High average turn duration: %s (threshold: 2m, sample: %d)",
					avg.Truncate(time.Second), counted,
				))
			}
		}
	}

	return warnings
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *MultiAgentCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
