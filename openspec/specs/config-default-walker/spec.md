# config-default-walker Specification

## Purpose
TBD - created by archiving change config-bootstrap-regression-fixes. Update Purpose after archive.
## Requirements
### Requirement: Struct-recursive default walker using mapstructure tags
The `config` package SHALL provide a `WalkDefaults(cfg *Config) map[string]interface{}` function that recursively traverses the `DefaultConfig()` struct and produces a flat map of viper-compatible default keys and values. The walker SHALL use `mapstructure` struct tags to derive the dot-separated key path, matching how viper unmarshals config fields.

#### Scenario: Walker produces viper-compatible keys
- **WHEN** `WalkDefaults(DefaultConfig())` is called
- **THEN** the returned map SHALL contain keys using dot-separated paths derived from `mapstructure` struct tags (e.g., `agent.provider`, `memory.maxTokens`)

#### Scenario: Walker handles nested structs
- **WHEN** a config struct contains nested structs (e.g., `Agent.Memory`)
- **THEN** the walker SHALL recurse into each nested struct and produce composite keys (e.g., `agent.memory.maxTokens`)

### Requirement: time.Duration field handling
The walker SHALL correctly handle `time.Duration` fields by emitting the duration value as-is, preserving the Go duration type for viper defaults.

#### Scenario: Duration field is emitted correctly
- **WHEN** `DefaultConfig()` contains a `time.Duration` field with value `30 * time.Second`
- **THEN** the walker SHALL include it in the map as a `time.Duration` value, not as an integer or string

### Requirement: String-based custom type handling
The walker SHALL handle string-based custom types (e.g., `type Provider string`) by emitting the underlying string value.

#### Scenario: String-based enum type is emitted as string
- **WHEN** a field has type `Provider` with underlying type `string` and default value `"openai"`
- **THEN** the walker SHALL emit the value as the string `"openai"`

### Requirement: Pointer-to-bool field handling
The walker SHALL handle `*bool` fields. When the pointer is non-nil, the walker SHALL emit the pointed-to bool value. When nil, the walker SHALL skip the field.

#### Scenario: Non-nil *bool is emitted
- **WHEN** a `*bool` field points to `true`
- **THEN** the walker SHALL emit `true` as the default value

#### Scenario: Nil *bool is skipped
- **WHEN** a `*bool` field is nil in `DefaultConfig()`
- **THEN** the walker SHALL NOT emit an entry for that key

### Requirement: Slice field handling
The walker SHALL emit slice fields as slice values when they are non-nil and non-empty in `DefaultConfig()`.

#### Scenario: Non-empty slice is emitted
- **WHEN** a field is `[]string{"a", "b"}` in `DefaultConfig()`
- **THEN** the walker SHALL emit the slice value directly

#### Scenario: Nil slice is skipped
- **WHEN** a slice field is nil in `DefaultConfig()`
- **THEN** the walker SHALL NOT emit an entry for that key

### Requirement: Map field handling
The walker SHALL emit map fields as map values when they are non-nil and non-empty in `DefaultConfig()`.

#### Scenario: Non-empty map is emitted
- **WHEN** a field is `map[string]string{"key": "val"}` in `DefaultConfig()`
- **THEN** the walker SHALL emit the map value directly

#### Scenario: Nil map is skipped
- **WHEN** a map field is nil in `DefaultConfig()`
- **THEN** the walker SHALL NOT emit an entry for that key

### Requirement: Parity test validates 1:1 match with viper defaults
A test SHALL exist that validates the walker output matches the expected viper defaults exactly. The test MUST call `WalkDefaults(DefaultConfig())` and assert that every key in the returned map has the correct value, and no expected keys are missing.

#### Scenario: Parity test passes for full DefaultConfig
- **WHEN** the parity test runs
- **THEN** every key produced by `WalkDefaults(DefaultConfig())` SHALL match the corresponding value that would be set by manual `viper.SetDefault()` calls

#### Scenario: Parity test catches new config fields
- **WHEN** a developer adds a new field to the config struct with a `mapstructure` tag and a non-zero default
- **THEN** the parity test SHALL automatically include the new field without any manual update to a defaults list

