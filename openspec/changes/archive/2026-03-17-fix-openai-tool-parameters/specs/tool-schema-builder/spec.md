## MODIFIED Requirements

### Requirement: Schema output includes additionalProperties
The `SchemaBuilder.Build()` method SHALL include `"additionalProperties": false` in the output schema. The `buildInputSchema()` function SHALL set the `AdditionalProperties` field to the JSON Schema false schema pattern.

#### Scenario: SchemaBuilder output
- **WHEN** `SchemaBuilder.Build()` is called
- **THEN** the returned map SHALL contain `"additionalProperties": false`

#### Scenario: buildInputSchema output
- **WHEN** `buildInputSchema()` constructs a `jsonschema.Schema`
- **THEN** the `AdditionalProperties` field SHALL be set to the false schema (`{Not: {}}`)
