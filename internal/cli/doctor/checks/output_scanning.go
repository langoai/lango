package checks

import (
	"context"
	"strings"

	"github.com/langowarny/lango/internal/config"
	"github.com/langowarny/lango/internal/session"
)

// OutputScanningCheck validates output scanning and interceptor configuration.
type OutputScanningCheck struct{}

// Name returns the check name.
func (c *OutputScanningCheck) Name() string {
	return "Output Scanning"
}

// Run checks output scanning configuration and state.
func (c *OutputScanningCheck) Run(ctx context.Context, cfg *config.Config) Result {
	if cfg == nil {
		return Result{Name: c.Name(), Status: StatusSkip, Message: "Configuration not loaded"}
	}

	if !cfg.Security.Interceptor.Enabled {
		// Check if secrets exist despite interceptor being disabled
		if cfg.Session.DatabasePath != "" {
			store, err := session.NewEntStore(cfg.Session.DatabasePath)
			if err == nil {
				defer store.Close()

				count, err := store.Client().Secret.Query().Count(ctx)
				if err == nil && count > 0 {
					return Result{
						Name:   c.Name(),
						Status: StatusWarn,
						Message: "Output interceptor is disabled but secrets exist in database",
						Details: "Stored secrets will not be redacted from AI output. " +
							"Enable security.interceptor.enabled to protect sensitive data.",
					}
				}
			} else {
				// DB might be encrypted — skip gracefully
				errMsg := err.Error()
				if strings.Contains(errMsg, "out of memory") ||
					strings.Contains(errMsg, "file is not a database") {
					return Result{
						Name:    c.Name(),
						Status:  StatusSkip,
						Message: "Cannot verify (database encrypted)",
					}
				}
			}
		}

		return Result{
			Name:    c.Name(),
			Status:  StatusSkip,
			Message: "Output interceptor is disabled",
		}
	}

	// Interceptor enabled
	if !cfg.Security.Interceptor.RedactPII {
		return Result{
			Name:    c.Name(),
			Status:  StatusWarn,
			Message: "Output interceptor enabled but PII redaction is disabled",
			Details: "Enable security.interceptor.redactPii for comprehensive output scanning.",
		}
	}

	return Result{
		Name:    c.Name(),
		Status:  StatusPass,
		Message: "Output scanning fully configured",
	}
}

// Fix delegates to Run as automatic fixing is not supported.
func (c *OutputScanningCheck) Fix(ctx context.Context, cfg *config.Config) Result {
	return c.Run(ctx, cfg)
}
