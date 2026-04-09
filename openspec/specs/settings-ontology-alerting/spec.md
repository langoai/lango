## Purpose

Capability spec for settings-ontology-alerting. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Ontology settings form with full config coverage
The settings TUI SHALL provide an "Ontology" category under the "AI & Knowledge" section with fields covering `OntologyConfig`, `OntologyACLConfig`, `OntologyGovernanceConfig`, and `OntologyExchangeConfig`. Sub-section fields SHALL be conditionally visible based on their parent enabled toggle.

#### Scenario: Ontology form created
- **WHEN** `createFormForCategory("ontology", cfg)` is called
- **THEN** the function SHALL return a non-nil `*tuicore.FormModel` with fields for all ontology config keys

#### Scenario: Ontology values saved
- **WHEN** user edits ontology fields and saves
- **THEN** `UpdateConfigFromForm` SHALL map all `ontology_*` field keys to `cfg.Ontology.*` paths, including `ontology_acl_roles` via `parseKeyValuePairs`

#### Scenario: ACL roles CSV editing
- **WHEN** user enters `"operator=write,librarian=read"` in the `ontology_acl_roles` field
- **THEN** `UpdateConfigFromForm` SHALL set `cfg.Ontology.ACL.Roles` to `map[string]string{"operator": "write", "librarian": "read"}`

### Requirement: Alerting settings form with observability dependency
The settings TUI SHALL provide an "Alerting" category under the "Integrations" section with 3 fields: enabled, policyBlockRateThreshold, recoveryRetryThreshold.

#### Scenario: Alerting form created
- **WHEN** `createFormForCategory("alerting", cfg)` is called
- **THEN** the function SHALL return a non-nil `*tuicore.FormModel`

#### Scenario: Alerting enabled check
- **WHEN** `categoryIsEnabled("alerting")` is evaluated
- **THEN** the result SHALL be `cfg.Observability.Enabled && cfg.Alerting.Enabled`

#### Scenario: Alerting dependency panel
- **WHEN** user navigates to the alerting category with `observability.enabled = false`
- **THEN** the dependency panel SHALL show "Observability" as an unmet required dependency
