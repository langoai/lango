package config

import "time"

// RunLedgerConfig defines the Task OS / RunLedger configuration.
type RunLedgerConfig struct {
	// Enabled activates the RunLedger system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Shadow mode: journal records only, existing systems unaffected.
	Shadow bool `mapstructure:"shadow" json:"shadow"`

	// WriteThrough: all creates/updates go through ledger first, then mirror.
	WriteThrough bool `mapstructure:"writeThrough" json:"writeThrough"`

	// AuthoritativeRead: state reads come from ledger snapshots only.
	AuthoritativeRead bool `mapstructure:"authoritativeRead" json:"authoritativeRead"`

	// StaleTTL is how long a paused run remains resumable (default: 1h).
	StaleTTL time.Duration `mapstructure:"staleTtl" json:"staleTtl"`

	// MaxRunHistory is the maximum number of runs to keep (0 = unlimited).
	MaxRunHistory int `mapstructure:"maxRunHistory" json:"maxRunHistory"`

	// ValidatorTimeout is the timeout for individual validator execution (default: 2m).
	ValidatorTimeout time.Duration `mapstructure:"validatorTimeout" json:"validatorTimeout"`

	// PlannerMaxRetries is how many times a malformed planner output is retried (default: 2).
	PlannerMaxRetries int `mapstructure:"plannerMaxRetries" json:"plannerMaxRetries"`
}
