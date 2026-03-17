## MODIFIED Requirements

### Requirement: Tool parameter schema extraction
The system SHALL extract tool parameter schemas from `genai.FunctionDeclaration` by checking `ParametersJsonSchema` first, then falling back to `Parameters`. When both fields are set, `ParametersJsonSchema` SHALL take priority.

#### Scenario: ADK v0.5.0+ tool with ParametersJsonSchema
- **WHEN** a `FunctionDeclaration` has `ParametersJsonSchema` set and `Parameters` nil
- **THEN** the system SHALL use `ParametersJsonSchema` to build the tool's parameter schema

#### Scenario: Legacy tool with Parameters only
- **WHEN** a `FunctionDeclaration` has `Parameters` set and `ParametersJsonSchema` nil
- **THEN** the system SHALL use `Parameters` to build the tool's parameter schema

#### Scenario: Both fields set
- **WHEN** a `FunctionDeclaration` has both `ParametersJsonSchema` and `Parameters` set
- **THEN** the system SHALL use `ParametersJsonSchema` and ignore `Parameters`

## ADDED Requirements

### Requirement: OpenAI strict mode for fully-required tools
The system SHALL set `Strict: true` on OpenAI `FunctionDefinition` when all declared properties are listed in `required` and `additionalProperties` is `false`.

#### Scenario: All properties required with additionalProperties false
- **WHEN** a tool schema has `additionalProperties: false` and every property is in the `required` array
- **THEN** the system SHALL set `Strict: true` on the OpenAI function definition

#### Scenario: Optional property exists
- **WHEN** a tool schema has `additionalProperties: false` but at least one property is not in `required`
- **THEN** the system SHALL set `Strict: false`

#### Scenario: No additionalProperties field
- **WHEN** a tool schema does not include `additionalProperties`
- **THEN** the system SHALL set `Strict: false`
