## Purpose

Capability spec for context-auto-enable. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: collectExplicitKeys
The system SHALL provide `collectExplicitKeys(configPath string, keys []string) map[string]bool` that reads the raw config file (JSON) and returns which of the given dotted keys are present. SHALL return nil if the file cannot be read.

#### Scenario: Keys present in config file
- **WHEN** config file contains `{"knowledge":{"enabled":false}}`
- **THEN** `collectExplicitKeys` SHALL return `map["knowledge.enabled": true]`

#### Scenario: Keys absent
- **WHEN** config file does not contain a key
- **THEN** that key SHALL NOT appear in the returned map

#### Scenario: File not found
- **WHEN** configPath is empty or file does not exist
- **THEN** SHALL return nil

### Requirement: ResolveContextAutoEnable
The system SHALL provide `ResolveContextAutoEnable(cfg *Config, explicitKeys map[string]bool) AutoEnabledSet` that auto-enables context subsystems when their config-level dependencies are detectable and the user has not explicitly disabled them.

#### Scenario: Knowledge auto-enable
- **WHEN** `explicitKeys["knowledge.enabled"]` is false AND `cfg.Session.DatabasePath != ""`
- **THEN** `cfg.Knowledge.Enabled` SHALL be set to true AND `AutoEnabledSet.Knowledge` SHALL be true

#### Scenario: Explicit disable respected
- **WHEN** `explicitKeys["knowledge.enabled"]` is true AND `cfg.Knowledge.Enabled` is false
- **THEN** Knowledge SHALL NOT be auto-enabled

#### Scenario: Retrieval follows Knowledge
- **WHEN** Knowledge will be enabled AND retrieval not explicitly disabled
- **THEN** `cfg.Retrieval.Enabled` SHALL be set to true

#### Scenario: nil explicitKeys
- **WHEN** explicitKeys is nil (legacy profile)
- **THEN** auto-enable SHALL treat all features as eligible (intentional migration)

### Requirement: ProbeEmbeddingProvider
The system SHALL provide `ProbeEmbeddingProvider() string` that scans `cfg.Providers` for embedding-capable providers. Policy: local-first (Ollama preferred), single-remote-only, multiple-remote → no auto-select.

#### Scenario: Single local provider
- **WHEN** only Ollama is configured
- **THEN** SHALL return the Ollama provider key

#### Scenario: Multiple remote providers
- **WHEN** both OpenAI and Gemini are configured (no local)
- **THEN** SHALL return "" (no auto-select to prevent cost surprise)

#### Scenario: Local + remote
- **WHEN** both Ollama and OpenAI are configured
- **THEN** SHALL return the Ollama provider key (local-first)

### Requirement: AutoEnabledSet
The system SHALL provide `AutoEnabledSet{Knowledge, Memory, Retrieval, Embedding bool}` recording which features were auto-enabled.

#### Scenario: AutoEnabledSet populated
- **WHEN** ResolveContextAutoEnable auto-enables Knowledge
- **THEN** `AutoEnabledSet.Knowledge` SHALL be true

### Requirement: PresetExplicitKeys
The system SHALL provide `PresetExplicitKeys(name string) map[string]bool` returning which context-related keys a preset explicitly sets.

#### Scenario: Researcher preset
- **WHEN** `PresetExplicitKeys("researcher")` is called
- **THEN** SHALL include "knowledge.enabled", "embedding.provider", "librarian.enabled"

#### Scenario: Minimal preset
- **WHEN** `PresetExplicitKeys("minimal")` is called
- **THEN** SHALL return an empty map
