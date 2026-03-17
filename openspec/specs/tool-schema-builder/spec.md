# tool-schema-builder Specification

## Purpose
TBD - created by archiving change config-bootstrap-regression-fixes. Update Purpose after archive.
## Requirements
### Requirement: Type-safe JSON Schema builder
The system SHALL provide a `SchemaBuilder` type that constructs JSON Schema objects in a type-safe, fluent manner. The builder SHALL support common JSON Schema types and constraints without requiring callers to manually assemble `map[string]interface{}` structures.

#### Scenario: Builder is instantiated
- **WHEN** a new `SchemaBuilder` is created
- **THEN** it SHALL initialize with `type: "object"` as the root schema type

### Requirement: Str method for string properties
The builder SHALL provide a `Str(name, description string)` method that adds a string-typed property to the schema.

#### Scenario: String property is added
- **WHEN** `builder.Str("name", "The user name")` is called
- **THEN** the resulting schema SHALL contain a property `name` with `type: "string"` and the given description

### Requirement: Int method for integer properties
The builder SHALL provide an `Int(name, description string)` method that adds an integer-typed property to the schema.

#### Scenario: Integer property is added
- **WHEN** `builder.Int("count", "Number of items")` is called
- **THEN** the resulting schema SHALL contain a property `count` with `type: "integer"` and the given description

### Requirement: Bool method for boolean properties
The builder SHALL provide a `Bool(name, description string)` method that adds a boolean-typed property to the schema.

#### Scenario: Boolean property is added
- **WHEN** `builder.Bool("verbose", "Enable verbose output")` is called
- **THEN** the resulting schema SHALL contain a property `verbose` with `type: "boolean"` and the given description

### Requirement: Enum method for enumerated string properties
The builder SHALL provide an `Enum(name, description string, values ...string)` method that adds a string property constrained to the given set of values.

#### Scenario: Enum property is added
- **WHEN** `builder.Enum("status", "Current status", "active", "inactive", "pending")` is called
- **THEN** the resulting schema SHALL contain a property `status` with `type: "string"`, the given description, and an `enum` array containing `["active", "inactive", "pending"]`

### Requirement: Required method marks mandatory properties
The builder SHALL provide a `Required(names ...string)` method that marks one or more properties as required in the schema.

#### Scenario: Required properties are set
- **WHEN** `builder.Required("name", "count")` is called
- **THEN** the resulting schema SHALL contain a `required` array with `["name", "count"]`

### Requirement: Build returns map compatible with agent.Tool.Parameters
The `Build()` method SHALL return a `map[string]interface{}` that is directly assignable to `agent.Tool.Parameters`. The structure MUST conform to JSON Schema draft-07 or later. The ADK `buildInputSchema()` function SHALL correctly consume this format by detecting the `"properties"` key and extracting nested property definitions and the top-level `required` array.

#### Scenario: Build produces valid schema map
- **WHEN** `builder.Str("query", "Search query").Required("query").Build()` is called
- **THEN** the returned map SHALL have keys `type`, `properties`, `additionalProperties`, and `required`
- **THEN** `properties` SHALL contain the `query` property definition
- **THEN** `required` SHALL be `["query"]`

### Requirement: Schema output includes additionalProperties
The `SchemaBuilder.Build()` method SHALL include `"additionalProperties": false` in the output schema. The `buildInputSchema()` function SHALL set the `AdditionalProperties` field to the JSON Schema false schema pattern.

#### Scenario: SchemaBuilder output
- **WHEN** `SchemaBuilder.Build()` is called
- **THEN** the returned map SHALL contain `"additionalProperties": false`

#### Scenario: buildInputSchema output
- **WHEN** `buildInputSchema()` constructs a `jsonschema.Schema`
- **THEN** the `AdditionalProperties` field SHALL be set to the false schema (`{Not: {}}`)

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

### Requirement: No runtime coupling with toolparam
The `SchemaBuilder` SHALL NOT import or depend on the `toolparam` package at runtime. The builder is a standalone schema construction utility. Any migration from `toolparam` to the builder is a caller-side concern.

#### Scenario: Builder has no toolparam import
- **WHEN** the builder package's import graph is inspected
- **THEN** it SHALL NOT contain any import of `toolparam` or equivalent parameter-binding packages

### Requirement: Fluent method chaining
All builder methods SHALL return the builder instance to support fluent method chaining.

#### Scenario: Methods are chainable
- **WHEN** `builder.Str("a", "desc").Int("b", "desc").Required("a").Build()` is called
- **THEN** the chain SHALL compile and produce a valid schema containing both properties with `a` required

