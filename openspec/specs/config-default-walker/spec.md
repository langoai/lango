# config-default-walker Specification

## Purpose
Recursively walk `DefaultConfig()` using mapstructure tags and register viper defaults directly — ensuring `DefaultConfig()` is the single source of truth for all default values.

## Requirements
### Requirement: Struct-recursive default walker using mapstructure tags
The `config` package SHALL provide an unexported `setDefaultsFromStruct(v *viper.Viper, prefix string, val reflect.Value)` function that recursively traverses the `DefaultConfig()` struct and sets viper defaults directly via `v.SetDefault()`. The walker SHALL use `mapstructure` struct tags to derive the dot-separated key path, matching how viper unmarshals config fields.

#### Scenario: Walker sets viper-compatible defaults
- **WHEN** `setDefaultsFromStruct(v, "", reflect.ValueOf(DefaultConfig()).Elem())` is called
- **THEN** each non-zero leaf field SHALL be registered via `v.SetDefault(key, value)` using dot-separated paths derived from `mapstructure` struct tags (e.g., `agent.provider`, `memory.maxTokens`)

#### Scenario: Walker handles nested structs
- **WHEN** a config struct contains nested structs (e.g., `Agent.Memory`)
- **THEN** the walker SHALL recurse into each nested struct and produce composite keys (e.g., `agent.memory.maxTokens`)

### Requirement: time.Duration field handling
The walker SHALL correctly handle `time.Duration` fields by setting the duration value as-is via `v.SetDefault()`, preserving the Go duration type.

#### Scenario: Duration field is set correctly
- **WHEN** `DefaultConfig()` contains a `time.Duration` field with value `30 * time.Second`
- **THEN** the walker SHALL call `v.SetDefault(key, 30*time.Second)` with the `time.Duration` value, not an integer or string

### Requirement: String-based custom type handling
The walker SHALL handle string-based custom types (e.g., `type Provider string`) by converting to the underlying string value via `actual.String()`.

#### Scenario: String-based enum type is set as string
- **WHEN** a field has type `Provider` with underlying type `string` and default value `"openai"`
- **THEN** the walker SHALL call `v.SetDefault()` with the plain string `"openai"`

### Requirement: Pointer-to-bool field handling
The walker SHALL handle `*bool` fields. When the pointer is non-nil, the walker SHALL dereference and set the pointed-to bool value. When nil, the walker SHALL skip the field.

#### Scenario: Non-nil *bool is set
- **WHEN** a `*bool` field points to `true`
- **THEN** the walker SHALL call `v.SetDefault(key, true)`

#### Scenario: Nil *bool is skipped
- **WHEN** a `*bool` field is nil in `DefaultConfig()`
- **THEN** the walker SHALL NOT call `v.SetDefault()` for that key

### Requirement: Slice field handling
The walker SHALL set non-nil, non-empty slice fields via `v.SetDefault()`.

#### Scenario: Non-empty slice is set
- **WHEN** a field is `[]string{"a", "b"}` in `DefaultConfig()`
- **THEN** the walker SHALL call `v.SetDefault(key, []string{"a", "b"})`

#### Scenario: Nil or empty slice is skipped
- **WHEN** a slice field is nil or has length 0 in `DefaultConfig()`
- **THEN** the walker SHALL NOT call `v.SetDefault()` for that key

### Requirement: Map field handling
The walker SHALL skip map fields entirely. Maps contain dynamic user content (e.g., provider definitions, server configs), not static defaults.

#### Scenario: Map field is skipped
- **WHEN** a field is `map[string]ProviderConfig{...}` in `DefaultConfig()`
- **THEN** the walker SHALL skip the field without calling `v.SetDefault()`

### Requirement: Zero-value field handling
The walker SHALL skip zero-value leaf fields to avoid overriding user-supplied values with empty defaults.

#### Scenario: Zero-value field is skipped
- **WHEN** a leaf field has its zero value (e.g., `""`, `0`, `false`)
- **THEN** the walker SHALL NOT call `v.SetDefault()` for that key

### Requirement: Parity test validates 1:1 match with viper defaults
A test SHALL exist that validates the walker output matches the expected viper defaults exactly. The test MUST load config via `Load("")` and assert that every key from `DefaultConfig()` has the correct value, and no expected keys are missing.

#### Scenario: Parity test passes for full DefaultConfig
- **WHEN** the parity test runs
- **THEN** every key produced by the walker SHALL match the corresponding value in the loaded config

#### Scenario: Parity test catches new config fields
- **WHEN** a developer adds a new field to the config struct with a `mapstructure` tag and a non-zero default
- **THEN** the parity test SHALL automatically include the new field without any manual update to a defaults list
