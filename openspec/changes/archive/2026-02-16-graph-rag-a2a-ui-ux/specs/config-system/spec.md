## ADDED Requirements

### Requirement: Graph config defaults
DefaultConfig SHALL include Graph defaults: Enabled=false, Backend="bolt", MaxTraversalDepth=2, MaxExpansionResults=10. Viper defaults SHALL be registered for these fields.

#### Scenario: New profile defaults
- **WHEN** a new profile is created via `lango config create`
- **THEN** graph config has Enabled=false, Backend="bolt", MaxTraversalDepth=2, MaxExpansionResults=10

### Requirement: A2A config defaults
DefaultConfig SHALL include A2A defaults: Enabled=false. Viper defaults SHALL be registered.

#### Scenario: New profile A2A defaults
- **WHEN** a new profile is created via `lango config create`
- **THEN** A2A config has Enabled=false

### Requirement: Graph config validation
Validate SHALL reject configurations where graph.enabled is true and graph.backend is not "bolt".

#### Scenario: Invalid graph backend
- **WHEN** config has graph.enabled=true and graph.backend="rocksdb"
- **THEN** Validate returns an error about unsupported backend

### Requirement: A2A config validation
Validate SHALL reject configurations where a2a.enabled is true but a2a.baseUrl or a2a.agentName is empty.

#### Scenario: A2A missing base URL
- **WHEN** config has a2a.enabled=true and a2a.baseUrl is empty
- **THEN** Validate returns an error about required baseUrl

#### Scenario: A2A missing agent name
- **WHEN** config has a2a.enabled=true and a2a.agentName is empty
- **THEN** Validate returns an error about required agentName
