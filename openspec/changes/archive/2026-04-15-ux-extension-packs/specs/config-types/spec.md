## ADDED Requirements

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
