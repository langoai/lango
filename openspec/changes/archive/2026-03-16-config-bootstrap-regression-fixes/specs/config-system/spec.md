## ADDED Requirements

### Requirement: PostLoad one-stop normalization and validation
The `config` package SHALL export a `PostLoad(*Config) error` function that applies all post-load processing in order: legacy migration, environment variable substitution, path normalization, path validation, and full config validation. All operations MUST be idempotent — calling PostLoad multiple times on the same config SHALL produce the same result.

#### Scenario: PostLoad applies full processing chain
- **WHEN** `PostLoad(cfg)` is called on a freshly deserialized config
- **THEN** the config has legacy fields migrated, env vars expanded, paths normalized to absolute, data paths validated under DataRoot, and full config validation applied

#### Scenario: PostLoad is idempotent
- **WHEN** `PostLoad(cfg)` is called twice on the same config
- **THEN** the second call produces no additional changes and returns the same result

## MODIFIED Requirements

### Requirement: Config loading applies normalization and validation
The `Load()` function SHALL delegate all post-load processing to `PostLoad()` instead of calling individual steps separately.

#### Scenario: Load delegates to PostLoad
- **WHEN** `config.Load(path)` is called
- **THEN** after unmarshalling, it calls `PostLoad(cfg)` once and returns the result
