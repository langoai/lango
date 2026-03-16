## MODIFIED Requirements

### Requirement: Build returns map compatible with agent.Tool.Parameters
The `Build()` method SHALL return a `map[string]interface{}` that is directly assignable to `agent.Tool.Parameters`. The structure MUST conform to JSON Schema draft-07 or later. The ADK `buildInputSchema()` function SHALL correctly consume this format by detecting the `"properties"` key and extracting nested property definitions and the top-level `required` array.

#### Scenario: Build produces valid schema map
- **WHEN** `builder.Str("query", "Search query").Required("query").Build()` is called
- **THEN** the returned map SHALL have keys `type`, `properties`, and `required`
- **THEN** `properties` SHALL contain the `query` property definition
- **THEN** `required` SHALL be `["query"]`

#### Scenario: Build output is assignable to agent.Tool
- **WHEN** the build output is assigned to `tool.Parameters`
- **THEN** it SHALL compile and function correctly without type assertion errors

#### Scenario: buildInputSchema correctly parses SchemaBuilder output
- **WHEN** `buildInputSchema()` receives a tool whose Parameters were set via `SchemaBuilder.Build()`
- **THEN** the resulting `jsonschema.Schema` SHALL contain only the actual parameter properties (not `type`, `properties`, or `required` as property names)
- **THEN** the `Required` field SHALL match the builder's `Required()` calls

#### Scenario: buildInputSchema preserves enum values from SchemaBuilder
- **WHEN** `buildInputSchema()` receives a tool with an Enum property from `SchemaBuilder.Build()`
- **THEN** the resulting schema property SHALL contain the enum values

## ADDED Requirements

### Requirement: buildInputSchema detects full JSON Schema format
The `buildInputSchema()` function SHALL detect when `agent.Tool.Parameters` contains a full JSON Schema object (with `"properties"` key mapping to a `map[string]interface{}`) and extract the nested property definitions instead of treating top-level keys as parameter names.

#### Scenario: Full JSON Schema format is detected
- **WHEN** `buildInputSchema()` receives Parameters with keys `type`, `properties`, and `required`
- **THEN** it SHALL use the value of `properties` as the parameter map
- **THEN** it SHALL extract `required` as a string slice from the top-level

#### Scenario: Flat ParameterDef format still works
- **WHEN** `buildInputSchema()` receives Parameters with `ParameterDef` values (no `properties` key)
- **THEN** it SHALL process them as before (type, description, required from ParameterDef fields)

#### Scenario: Flat map format still works
- **WHEN** `buildInputSchema()` receives Parameters with raw map values (no `properties` key)
- **THEN** it SHALL process them as before (type, description, required from map keys)
