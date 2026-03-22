## ADDED Requirements

### Requirement: Provenance Configuration Section
The Config struct SHALL include a `Provenance ProvenanceConfig` field with sub-struct `CheckpointConfig`. DefaultConfig SHALL set provenance defaults: enabled=false, autoOnStepComplete=true, autoOnPolicy=true, maxPerSession=100, retentionDays=30.

#### Scenario: Default config includes provenance
- **WHEN** DefaultConfig() is called
- **THEN** the Provenance field is populated with default values
