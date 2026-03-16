## Why

`buildInputSchema()` in `internal/adk/tools.go` treats top-level keys of `agent.Tool.Parameters` as parameter names. When `SchemaBuilder.Build()` produces a full JSON Schema (`{"type":"object","properties":{...},"required":[...]}`), the function misinterprets `type`, `properties`, and `required` as parameter names instead of parsing the nested structure. This causes all SchemaBuilder-based tools (exec, filesystem, browser, output, security) to register with incorrect schemas, leading to "missing command parameter" errors at runtime.

## What Changes

- Fix `buildInputSchema()` to detect full JSON Schema format (presence of `"properties"` key containing a map) and extract nested properties and top-level `required` array
- Add enum extraction support for map-based parameter definitions (SchemaBuilder stores enums in the nested property maps)
- Add comprehensive tests for SchemaBuilder format, flat ParameterDef format, and flat map format

## Capabilities

### New Capabilities

### Modified Capabilities
- `tool-schema-builder`: buildInputSchema() must correctly consume SchemaBuilder.Build() output, extracting nested properties and required arrays instead of treating top-level JSON Schema keys as parameter names

## Impact

- `internal/adk/tools.go` — `buildInputSchema()` function
- `internal/adk/tools_test.go` — new test cases
- All tools using `SchemaBuilder.Build()` are unblocked (exec, filesystem, browser, output, security tool files)
