## Purpose

Capability spec for config-types. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: ProviderConfig type strengthening
The `ProviderConfig.Type` field SHALL use `types.ProviderType` instead of raw `string`.

#### Scenario: Config deserialization with typed provider
- **WHEN** config is loaded via mapstructure/viper
- **THEN** `ProviderConfig.Type` SHALL deserialize correctly as `types.ProviderType`

#### Scenario: Provider validation
- **WHEN** a `ProviderConfig` is created with an unknown provider type
- **THEN** `config.Type.Valid()` SHALL return `false`

### Requirement: AgentConfig fields
`AgentConfig` SHALL include `MaxTurns int`, `ErrorCorrectionEnabled *bool`, and `MaxDelegationRounds int` fields with mapstructure/json tags.

#### Scenario: Zero-value defaults
- **WHEN** config omits `maxTurns`, `errorCorrectionEnabled`, and `maxDelegationRounds`
- **THEN** the zero values SHALL be interpreted as defaults by the wiring layer
- **AND** the effective defaults SHALL be 50 turns in single-agent mode, 75 turns in multi-agent mode, true for error correction, and 10 for max delegation rounds

### Requirement: ObservationalMemoryConfig fields
`ObservationalMemoryConfig` SHALL include `MemoryTokenBudget int` and `ReflectionConsolidationThreshold int` fields with mapstructure/json tags.

#### Scenario: Zero-value defaults
- **WHEN** config omits `memoryTokenBudget` and `reflectionConsolidationThreshold`
- **THEN** the zero values SHALL be interpreted as defaults (4000, 5) by the wiring layer

### Requirement: Economy configuration struct
The config package SHALL include an EconomyConfig struct with sub-configs for all 5 subsystems. The struct SHALL use mapstructure tags for viper binding.

#### Scenario: Economy config loaded
- **WHEN** configuration is loaded with economy section
- **THEN** EconomyConfig is populated with Budget, Risk, Negotiate, Escrow, and Pricing sub-configs

### Requirement: Config field in main config
The main Config struct SHALL include an Economy field of type EconomyConfig, enabling `economy.enabled`, `economy.budget.*`, etc. configuration paths.

#### Scenario: Economy disabled by default
- **WHEN** no economy config is provided
- **THEN** economy.enabled defaults to false

### Requirement: ExtensionsConfig struct
The config package SHALL include an `ExtensionsConfig` struct in `internal/config/types.go` with fields `Enabled *bool` (default `true`), `Dir string` (default `~/.lango/extensions`), and `EnforceIntegrity bool` (default `false`). The struct SHALL use `mapstructure` and `json` tags matching the field names in lower-camel case. `Config` SHALL carry an `Extensions ExtensionsConfig` field.

#### Scenario: Default values applied
- **WHEN** the config loader runs with no `extensions.*` keys set
- **THEN** `Config.Extensions.Enabled` SHALL be a non-nil pointer to `true`
- **AND** `Config.Extensions.Dir` SHALL be `~/.lango/extensions` (or its expanded form)
- **AND** `Config.Extensions.EnforceIntegrity` SHALL be `false`

#### Scenario: User override preserved
- **WHEN** the user sets `extensions.enabled: false` and `extensions.dir: /data/packs`
- **THEN** the loaded `Extensions` config SHALL reflect those values

#### Scenario: Tilde expansion for Dir
- **WHEN** `extensions.dir` contains a leading `~/`
- **THEN** resolution to an absolute path SHALL expand `~/` to the current user's home directory at consumption time (not at load time)

### Requirement: ResolveExtensions accessor
The config package SHALL expose a `(ExtensionsConfig) ResolveExtensions() ExtensionsConfig` method that returns a copy with defaults applied (non-nil `Enabled`, non-empty `Dir`). The method SHALL NOT mutate the receiver.

#### Scenario: Resolve fills missing fields
- **WHEN** `ResolveExtensions` is called on an empty `ExtensionsConfig{}`
- **THEN** the returned struct SHALL have `Enabled=*true`, `Dir="~/.lango/extensions"`, and `EnforceIntegrity=false`

#### Scenario: Receiver not mutated
- **WHEN** `ResolveExtensions` is called on a non-empty struct
- **THEN** the original struct SHALL be unchanged after the call

