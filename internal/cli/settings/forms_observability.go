package settings

import (
	"fmt"
	"strconv"
	"time"

	"github.com/langoai/lango/internal/cli/tuicore"
	"github.com/langoai/lango/internal/config"
)

// NewObservabilityForm creates the Observability configuration form.
func NewObservabilityForm(cfg *config.Config) *tuicore.FormModel {
	form := tuicore.NewFormModel("Observability Configuration")

	obsEnabledField := &tuicore.Field{
		Key: "obs_enabled", Label: "Enabled", Type: tuicore.InputBool,
		Checked:     cfg.Observability.Enabled,
		Description: "Enable the observability subsystem (metrics, tokens, health, audit)",
	}
	form.AddField(obsEnabledField)

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

	// --- Trace Store ---

	isObsEnabled := func() bool { return obsEnabledField.Checked }

	ts := cfg.Observability.TraceStore
	tsDefaults := config.TraceStoreDefaults()

	tsMaxAge := ts.MaxAge
	if tsMaxAge == 0 {
		tsMaxAge = tsDefaults.MaxAge
	}
	form.AddField(&tuicore.Field{
		Key: "obs_trace_max_age", Label: "Trace Max Age", Type: tuicore.InputText,
		Value:       tsMaxAge.String(),
		Placeholder: "720h",
		Description: "Maximum age of traces before cleanup (e.g. 720h for 30 days)",
		VisibleWhen: isObsEnabled,
		Validate: func(s string) error {
			if _, err := time.ParseDuration(s); err != nil {
				return fmt.Errorf("must be a valid duration (e.g. 720h, 168h)")
			}
			return nil
		},
	})

	tsMaxTraces := ts.MaxTraces
	if tsMaxTraces == 0 {
		tsMaxTraces = tsDefaults.MaxTraces
	}
	form.AddField(&tuicore.Field{
		Key: "obs_trace_max_traces", Label: "  Trace Max Count", Type: tuicore.InputInt,
		Value:       strconv.Itoa(tsMaxTraces),
		Placeholder: "10000",
		Description: "Maximum number of traces to retain",
		VisibleWhen: isObsEnabled,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	tsFailedMult := ts.FailedTraceMultiplier
	if tsFailedMult == 0 {
		tsFailedMult = tsDefaults.FailedTraceMultiplier
	}
	form.AddField(&tuicore.Field{
		Key: "obs_trace_failed_multiplier", Label: "  Failed Trace Multiplier", Type: tuicore.InputInt,
		Value:       strconv.Itoa(tsFailedMult),
		Placeholder: "2",
		Description: "Retention multiplier for failed traces (e.g. 2 = keep 2x longer)",
		VisibleWhen: isObsEnabled,
		Validate: func(s string) error {
			if i, err := strconv.Atoi(s); err != nil || i <= 0 {
				return fmt.Errorf("must be a positive integer")
			}
			return nil
		},
	})

	tsCleanup := ts.CleanupInterval
	if tsCleanup == 0 {
		tsCleanup = tsDefaults.CleanupInterval
	}
	form.AddField(&tuicore.Field{
		Key: "obs_trace_cleanup_interval", Label: "  Trace Cleanup Interval", Type: tuicore.InputText,
		Value:       tsCleanup.String(),
		Placeholder: "1h",
		Description: "How often the cleanup goroutine runs",
		VisibleWhen: isObsEnabled,
		Validate: func(s string) error {
			if _, err := time.ParseDuration(s); err != nil {
				return fmt.Errorf("must be a valid duration (e.g. 1h, 30m)")
			}
			return nil
		},
	})

	return &form
}
