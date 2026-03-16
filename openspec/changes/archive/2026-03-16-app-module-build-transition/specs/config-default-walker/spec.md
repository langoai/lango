## MODIFIED Requirements

### Requirement: Struct-recursive default walker using mapstructure tags
The `config` package SHALL provide an unexported `setDefaultsFromStruct(v *viper.Viper, prefix string, val reflect.Value)` function that recursively traverses the `DefaultConfig()` struct and sets viper defaults directly via `v.SetDefault()`. The walker SHALL use `mapstructure` struct tags to derive the dot-separated key path, matching how viper unmarshals config fields.

#### Scenario: Walker sets viper-compatible defaults
- **WHEN** `setDefaultsFromStruct(v, "", reflect.ValueOf(DefaultConfig()).Elem())` is called
- **THEN** each non-zero leaf field SHALL be registered via `v.SetDefault(key, value)` using dot-separated paths derived from `mapstructure` struct tags

#### Scenario: Walker handles nested structs
- **WHEN** a config struct contains nested structs
- **THEN** the walker SHALL recurse into each nested struct and produce composite keys

### Requirement: Map field handling
The walker SHALL skip map fields entirely. Maps contain dynamic user content (e.g., provider definitions, server configs), not static defaults.

#### Scenario: Map field is skipped
- **WHEN** a field is `map[string]ProviderConfig{...}` in `DefaultConfig()`
- **THEN** the walker SHALL skip the field without calling `v.SetDefault()`
