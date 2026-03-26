## MODIFIED Requirements

### Requirement: Config loading returns structured result
`config.Load()` SHALL return `(*LoadResult, error)` instead of `(*Config, error)`. `LoadResult` SHALL contain `Config *Config` and `ExplicitKeys map[string]bool`. All existing callers of `config.Load()` MUST be updated to use `result.Config` for the config object.

#### Scenario: Existing callers compile after signature change
- **WHEN** `config.Load()` is called from `cmd/lango/main.go`, `internal/bootstrap/bootstrap.go`, `internal/configstore/migrate.go`, or `internal/cli/doctor/doctor.go`
- **THEN** each caller accesses `result.Config` and the project builds without errors

#### Scenario: ExplicitKeys collected from raw viper
- **WHEN** config file sets `knowledge.enabled: true` and `librarian.enabled: false`
- **THEN** `LoadResult.ExplicitKeys` contains both keys, and does NOT contain keys only present via `SetDefault()`

### Requirement: ContextProfile field in Config
`Config` SHALL have a `ContextProfile ContextProfileName` field with mapstructure tag `contextProfile`. The field SHALL accept values `off`, `lite`, `balanced`, `full`, or empty string (no profile).

#### Scenario: ContextProfile unmarshaled from JSON config
- **WHEN** config file contains `"contextProfile": "balanced"`
- **THEN** `cfg.ContextProfile` equals `ContextProfileBalanced`

#### Scenario: Empty profile means no profile applied
- **WHEN** config file does not set `contextProfile`
- **THEN** `cfg.ContextProfile` is empty string and `ApplyContextProfile` is a no-op

### Requirement: ApplyContextProfile in load pipeline
`ApplyContextProfile(cfg, explicitKeys)` SHALL be called inside `Load()` after `Unmarshal` and before `PostLoad`. `PostLoad`'s signature SHALL NOT change.

#### Scenario: Profile applied before validation
- **WHEN** `contextProfile: balanced` is set
- **THEN** `Knowledge.Enabled` is `true` before `Validate()` runs, so downstream validation sees the profile-applied state
