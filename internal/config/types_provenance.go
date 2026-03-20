package config

// ProvenanceConfig defines session provenance tracking configuration.
type ProvenanceConfig struct {
	// Enabled activates the provenance system.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// Checkpoints configures automatic checkpoint creation.
	Checkpoints CheckpointConfig `mapstructure:"checkpoints" json:"checkpoints"`
}

// CheckpointConfig defines checkpoint behavior.
type CheckpointConfig struct {
	// AutoOnStepComplete creates a checkpoint when a RunLedger step passes validation.
	AutoOnStepComplete bool `mapstructure:"autoOnStepComplete" json:"autoOnStepComplete"`

	// AutoOnPolicy creates a checkpoint when a policy decision is applied.
	AutoOnPolicy bool `mapstructure:"autoOnPolicy" json:"autoOnPolicy"`

	// MaxPerSession limits the number of checkpoints per session (0 = unlimited).
	MaxPerSession int `mapstructure:"maxPerSession" json:"maxPerSession"`

	// RetentionDays is how long checkpoints are kept before pruning (0 = unlimited).
	RetentionDays int `mapstructure:"retentionDays" json:"retentionDays"`
}
