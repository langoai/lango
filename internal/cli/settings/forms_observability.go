package settings

import (
	"fmt"
	"strconv"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewObservabilityForm creates the Observability configuration form.
func NewObservabilityForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Observability Configuration")

	form.AddField(&tuicore.Field{
		Key: "obs_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Enabled,
		Description: "Enable the observability subsystem (metrics, tokens, health, audit)",
	})

	// Token Tracking
	form.AddField(&tuicore.Field{
		Key: "obs_tokens_enabled", Label: "Token Tracking", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Tokens.Enabled,
		Description: "Track token usage per session, agent, and tool",
	})

	form.AddField(&tuicore.Field{
		Key: "obs_tokens_persist", Label: "  Persist History", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Tokens.PersistHistory,
		Description: "Store token usage records in the database",
	})

	form.AddField(&tuicore.Field{
		Key: "obs_tokens_retention", Label: "  Retention Days", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Observability.Tokens.RetentionDays),
		Placeholder: "30",
		Description: "Days to retain token usage records",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	// Health Checks
	form.AddField(&tuicore.Field{
		Key: "obs_health_enabled", Label: "Health Checks", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Health.Enabled,
		Description: "Enable health check monitoring",
	})

	form.AddField(&tuicore.Field{
		Key: "obs_health_interval", Label: "  Check Interval", Type: tuicore.InputText,
		Value:       cfg.Observability.Health.Interval.String(),
		Placeholder: "30s",
		Description: "Interval between health check probes",
	})

	// Audit Logging
	form.AddField(&tuicore.Field{
		Key: "obs_audit_enabled", Label: "Audit Logging", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Audit.Enabled,
		Description: "Record audit logs for tool and token events",
	})

	form.AddField(&tuicore.Field{
		Key: "obs_audit_retention", Label: "  Retention Days", Type: tuicore.InputInt,
		Value:       strconv.Itoa(cfg.Observability.Audit.RetentionDays),
		Placeholder: "90",
		Description: "Days to retain audit log records",
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	// Metrics Export
	form.AddField(&tuicore.Field{
		Key: "obs_metrics_enabled", Label: "Metrics Export", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Metrics.Enabled,
		Description: "Enable metrics export endpoint",
	})

	metricsFormat := cfg.Observability.Metrics.Format
	if metricsFormat == "" {
		metricsFormat = "json"
	}
	form.AddField(&tuicore.Field{
		Key: "obs_metrics_format", Label: "  Export Format", Type: tuicore.InputSelect,
		Value:       metricsFormat,
		Options:     []string{"json", "prometheus"},
		Description: "Metrics export format for the /metrics endpoint",
	})

	return &form
}
