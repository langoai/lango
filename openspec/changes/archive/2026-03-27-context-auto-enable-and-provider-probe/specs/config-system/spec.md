## MODIFIED Requirements

### Requirement: Load return type
`config.Load()` SHALL return `(*LoadResult, error)` instead of `(*Config, error)`. `LoadResult` SHALL contain `Config *Config`, `ExplicitKeys map[string]bool`, and `AutoEnabled AutoEnabledSet`.

#### Scenario: Load returns LoadResult
- **WHEN** `config.Load(path)` is called
- **THEN** it SHALL return a `*LoadResult` with Config, ExplicitKeys, and AutoEnabled populated

#### Scenario: Load pipeline order
- **WHEN** `config.Load(path)` is called
- **THEN** it SHALL execute: Unmarshal → collectExplicitKeys → ApplyContextProfile → ResolveContextAutoEnable → PostLoad

### Requirement: Validate contextProfile
`Validate()` SHALL reject unknown contextProfile names with an error containing "invalid contextProfile".

#### Scenario: Invalid profile rejected
- **WHEN** `contextProfile: "turbo"` is set
- **THEN** `Validate()` SHALL return an error containing "invalid contextProfile"

## ADDED Requirements

### Requirement: configstore profilePayload
The configstore SHALL wrap Config and ExplicitKeys in a `profilePayload` struct stored inside the encrypted profile. `Save()` SHALL accept explicitKeys parameter. `Load()/LoadActive()` SHALL return explicitKeys alongside Config. Legacy profiles without ExplicitKeys SHALL return nil.

#### Scenario: Save with explicitKeys
- **WHEN** `Save(ctx, name, cfg, explicitKeys)` is called
- **THEN** both Config and ExplicitKeys SHALL be encrypted and stored together

#### Scenario: Load legacy profile
- **WHEN** a profile saved before Step 8 is loaded
- **THEN** Config SHALL be returned normally and ExplicitKeys SHALL be nil

### Requirement: Bootstrap carries ExplicitKeys and AutoEnabled
`bootstrap.Result` SHALL include `ExplicitKeys map[string]bool` and `AutoEnabled config.AutoEnabledSet`. `phaseLoadProfile` SHALL call `ApplyContextProfile` and `ResolveContextAutoEnable` after profile load.

#### Scenario: Bootstrap resolves auto-enable
- **WHEN** bootstrap loads a profile with DatabasePath configured
- **THEN** `Result.AutoEnabled` SHALL reflect auto-enabled features
