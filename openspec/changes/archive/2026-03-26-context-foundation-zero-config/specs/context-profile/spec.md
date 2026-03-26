## ADDED Requirements

### Requirement: Profile-based context configuration
The system SHALL support a `contextProfile` config field that applies named preset configurations for all context engineering subsystems. Valid profile names are `off`, `lite`, `balanced`, and `full`.

#### Scenario: Balanced profile enables knowledge, memory, and librarian
- **WHEN** user sets `contextProfile: balanced` in config
- **THEN** `Knowledge.Enabled` is `true`, `ObservationalMemory.Enabled` is `true`, `Librarian.Enabled` is `true`, `Graph.Enabled` is `false`, and `Embedding` settings are unchanged

#### Scenario: Full profile enables all context subsystems
- **WHEN** user sets `contextProfile: full` in config
- **THEN** `Knowledge.Enabled`, `ObservationalMemory.Enabled`, `Librarian.Enabled`, and `Graph.Enabled` are all `true`

#### Scenario: Off profile disables all context subsystems
- **WHEN** user sets `contextProfile: off` in config
- **THEN** `Knowledge.Enabled`, `ObservationalMemory.Enabled`, `Librarian.Enabled`, and `Graph.Enabled` are all `false`

#### Scenario: Lite profile enables only knowledge and memory
- **WHEN** user sets `contextProfile: lite` in config
- **THEN** `Knowledge.Enabled` is `true`, `ObservationalMemory.Enabled` is `true`, `Librarian.Enabled` is `false`, `Graph.Enabled` is `false`

### Requirement: User explicit overrides take precedence over profile
The system SHALL NOT overwrite a config field that the user explicitly set in their config file, even when a profile would set it to a different value. Explicit key detection MUST use a separate viper instance without defaults.

#### Scenario: Explicit false is preserved against balanced profile
- **WHEN** user sets `contextProfile: balanced` AND explicitly sets `knowledge.enabled: false` in config file
- **THEN** `Knowledge.Enabled` remains `false`

#### Scenario: Explicit true is preserved against off profile
- **WHEN** user sets `contextProfile: off` AND explicitly sets `graph.enabled: true` in config file
- **THEN** `Graph.Enabled` remains `true`

#### Scenario: No config file means no explicit overrides
- **WHEN** no config file exists AND `contextProfile: balanced` is set via environment or flag
- **THEN** all profile defaults apply without restriction (explicitKeys is nil)

### Requirement: Profile applies before validation
The system SHALL apply `ApplyContextProfile()` after config unmarshal but before `PostLoad()` validation, so that profile-set values are validated normally.

#### Scenario: Invalid profile name rejected
- **WHEN** user sets `contextProfile: turbo` (invalid name)
- **THEN** config validation returns an error mentioning valid profile names

### Requirement: LoadResult return type
`config.Load()` SHALL return `*LoadResult` containing both the `*Config` and `ExplicitKeys map[string]bool`. The `ExplicitKeys` map SHALL only contain keys that the user explicitly wrote in their config file.

#### Scenario: LoadResult provides explicit keys
- **WHEN** config file contains `knowledge.enabled: false` and `graph.enabled: true`
- **THEN** `LoadResult.ExplicitKeys` contains `{"knowledge.enabled": true, "graph.enabled": true}` and does NOT contain `"observationalMemory.enabled"`
