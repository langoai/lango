package config

// OntologyConfig configures the ontology subsystem.
type OntologyConfig struct {
	// Enabled activates the ontology registry and seed migration.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// ACL configures operation-level access control.
	ACL OntologyACLConfig `mapstructure:"acl" json:"acl,omitempty"`
	// Governance configures schema lifecycle management.
	Governance OntologyGovernanceConfig `mapstructure:"governance" json:"governance,omitempty"`
	// Exchange configures P2P ontology exchange.
	Exchange OntologyExchangeConfig `mapstructure:"exchange" json:"exchange,omitempty"`
}

// OntologyGovernanceConfig configures schema lifecycle governance.
type OntologyGovernanceConfig struct {
	// Enabled activates governance FSM enforcement on RegisterType/RegisterPredicate.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// MaxNewPerDay is the combined daily limit for type + predicate proposals.
	MaxNewPerDay int `mapstructure:"maxNewPerDay" json:"maxNewPerDay,omitempty"`
	// QuarantinePeriodHrs is the quarantine duration in hours.
	QuarantinePeriodHrs int `mapstructure:"quarantinePeriodHrs" json:"quarantinePeriodHrs,omitempty"`
	// ShadowModeDurationHrs is the shadow mode duration in hours.
	ShadowModeDurationHrs int `mapstructure:"shadowModeDurationHrs" json:"shadowModeDurationHrs,omitempty"`
	// MinUsageForPromotion is the minimum usage count for auto-promotion.
	MinUsageForPromotion int `mapstructure:"minUsageForPromotion" json:"minUsageForPromotion,omitempty"`
	// SchemaExplosionBudget is the monthly limit for new proposals.
	SchemaExplosionBudget int `mapstructure:"schemaExplosionBudget" json:"schemaExplosionBudget,omitempty"`
}

// OntologyACLConfig configures role-based access control for ontology operations.
type OntologyACLConfig struct {
	// Enabled activates ACL policy enforcement.
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// Roles maps principal names to permission levels ("read", "write", "admin").
	Roles map[string]string `mapstructure:"roles" json:"roles,omitempty"`
	// P2PPermission is the default permission for peer: prefix principals (default "read").
	P2PPermission string `mapstructure:"p2pPermission" json:"p2pPermission,omitempty"`
}

// OntologyExchangeConfig configures P2P ontology exchange behavior.
type OntologyExchangeConfig struct {
	// Enabled activates P2P ontology exchange (requires both P2P and Ontology enabled).
	Enabled bool `mapstructure:"enabled" json:"enabled,omitempty"`
	// MinTrustForSchema is the minimum peer trust score for schema exchange (default 0.5).
	MinTrustForSchema float64 `mapstructure:"minTrustForSchema" json:"minTrustForSchema,omitempty"`
	// MinTrustForFacts is the minimum peer trust score for fact exchange (default 0.7).
	MinTrustForFacts float64 `mapstructure:"minTrustForFacts" json:"minTrustForFacts,omitempty"`
	// AutoImportMode determines how proposed schemas are imported: "shadow" (default), "governed", "disabled".
	AutoImportMode string `mapstructure:"autoImportMode" json:"autoImportMode,omitempty"`
	// MaxTypesPerImport limits types imported from a single peer exchange (default 10).
	MaxTypesPerImport int `mapstructure:"maxTypesPerImport" json:"maxTypesPerImport,omitempty"`
}
