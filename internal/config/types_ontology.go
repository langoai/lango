package config

// OntologyConfig configures the ontology subsystem.
type OntologyConfig struct {
	// Enabled activates the ontology registry and seed migration.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// ACL configures operation-level access control.
	ACL OntologyACLConfig `mapstructure:"acl" json:"acl,omitempty"`
}

// OntologyACLConfig configures role-based access control for ontology operations.
type OntologyACLConfig struct {
	// Enabled activates ACL policy enforcement.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// Roles maps principal names to permission levels ("read", "write", "admin").
	Roles map[string]string `mapstructure:"roles" json:"roles,omitempty"`
}
