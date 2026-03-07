package checks

import (
	"context"
	"fmt"

	"github.com/langoai/lango/internal/config"
)

// ObservabilityCheck validates observability configuration.
type ObservabilityCheck struct{}

// Name returns the check name.
func (c *ObservabilityCheck) Name() string {
	return "Observability"
}

// Run checks observability configuration validity.
func (c *ObservabilityCheck) Run(_ context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Observability.Enabled {
		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Observability not enabled (observability.enabled = false)",
		}
	}

	var issues []string
	status := StatusPass

	// Validate token tracking retention.
	if cfg.Observability.Tokens.PersistHistory && cfg.Observability.Tokens.RetentionDays <= 0 {
		issues = append(issues, "tokens.retentionDays should be positive when persistHistory is enabled")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Validate health check interval.
	if cfg.Observability.Health.Enabled && cfg.Observability.Health.Interval <= 0 {
		issues = append(issues, "health.interval should be positive when health checks are enabled")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	// Validate audit retention.
	if cfg.Observability.Audit.Enabled && cfg.Observability.Audit.RetentionDays <= 0 {
		issues = append(issues, "audit.retentionDays should be positive when audit logging is enabled")
		if status < StatusWarn {
			status = StatusWarn
		}
	}

	if len(issues) == 0 {
		features := "tokens"
		if cfg.Observability.Health.Enabled {
			features += ", health"
		}
		if cfg.Observability.Audit.Enabled {
			features += ", audit"
		}
		if cfg.Observability.Metrics.Enabled {
			features += ", metrics"
		}
		return Result{
			Name:    c.Name(),
			Status:  StatusPass,
			Message: fmt.Sprintf("Observability configured (%s)", features),
		}
	}

	message := "Observability issues:\n"
	for _, issue := range issues {
		message += fmt.Sprintf("- %s\n", issue)
	}
	return Result{Name: c.Name(), Status: status, Message: message}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *ObservabilityCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
