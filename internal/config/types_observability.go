package config

import "time"

// ObservabilityConfig defines observability and monitoring settings.
type ObservabilityConfig struct {
	// Enabled activates the observability subsystem.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Tokens configures token usage tracking.
	Tokens TokenTrackingConfig `mapstructure:"tokens" json:"tokens"`

	// Health configures health check monitoring.
	Health HealthConfig `mapstructure:"health" json:"health"`

	// Audit configures audit log recording.
	Audit AuditConfig `mapstructure:"audit" json:"audit"`

	// Metrics configures metrics export format.
	Metrics MetricsExportConfig `mapstructure:"metrics" json:"metrics"`

	// TraceStore configures turn trace retention and cleanup.
	TraceStore TraceStoreConfig `mapstructure:"traceStore" json:"traceStore"`

	// Tracing configures OpenTelemetry distributed tracing.
	Tracing TracingConfig `mapstructure:"tracing" json:"tracing"`
}

// TracingConfig defines OpenTelemetry tracing settings.
type TracingConfig struct {
	// Enabled activates distributed tracing.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Exporter selects the trace exporter: "stdout" (default) or "none".
	Exporter string `mapstructure:"exporter" json:"exporter"`
}

// TokenTrackingConfig defines token usage tracking settings.
type TokenTrackingConfig struct {
	// Enabled activates token tracking (default: true when observability is enabled).
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// PersistHistory enables DB-backed persistent storage.
	PersistHistory bool `mapstructure:"persistHistory" json:"persistHistory"`

	// RetentionDays controls how long token usage records are kept (default: 30).
	RetentionDays int `mapstructure:"retentionDays" json:"retentionDays"`
}

// HealthConfig defines health check settings.
type HealthConfig struct {
	// Enabled activates health checks (default: true when observability is enabled).
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Interval is the health check interval (default: 30s).
	Interval time.Duration `mapstructure:"interval" json:"interval"`
}

// AuditConfig defines audit log settings.
type AuditConfig struct {
	// Enabled activates audit logging.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// RetentionDays controls how long audit records are kept (default: 90).
	RetentionDays int `mapstructure:"retentionDays" json:"retentionDays"`
}

// MetricsExportConfig defines metrics export settings.
type MetricsExportConfig struct {
	// Enabled activates metrics export endpoint.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Format is the metrics export format (default: "json").
	Format string `mapstructure:"format" json:"format"`
}
