package config

// OntologyConfig configures the ontology subsystem.
type OntologyConfig struct {
	// Enabled activates the ontology registry and seed migration.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
}
