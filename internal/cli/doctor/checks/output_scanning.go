package checks

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/langoai/lango/internal/config"
	"github.com/langoai/lango/internal/session"
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
						Name:    c.Name(),
						Status:  StatusWarn,
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

	// Check Presidio connectivity if enabled.
	if cfg.Security.Interceptor.Presidio.Enabled {
		url := cfg.Security.Interceptor.Presidio.URL
		if url == "" {
			url = "http://localhost:5002"
		}
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(url + "/health")
		if err != nil {
			return Result{
				Name:    c.Name(),
				Status:  StatusWarn,
				Message: "Presidio enabled but not reachable",
				Details: "Presidio analyzer at " + url + " is not responding. " +
					"Regex-based PII detection will still work. " +
					"Start Presidio with: docker compose --profile presidio up -d",
			}
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return Result{
				Name:    c.Name(),
				Status:  StatusWarn,
				Message: "Presidio health check returned non-OK status",
				Details: "Presidio at " + url + " returned status " + resp.Status,
			}
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
