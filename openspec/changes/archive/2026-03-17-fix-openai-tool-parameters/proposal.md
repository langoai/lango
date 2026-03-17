## Why

OpenAI tool calling fails because `convertTools()` reads only the legacy `Parameters` field from `genai.FunctionDeclaration`, while ADK v0.5.0+ sets `ParametersJsonSchema` instead. This causes all tools to be sent to OpenAI with empty parameter schemas (`{}`), making tool calls fail or produce wrong arguments.

## What Changes

- Fix `convertTools()` to prefer `ParametersJsonSchema` over `Parameters`, with fallback to legacy field
- Add `additionalProperties: false` to `SchemaBuilder.Build()` and `buildInputSchema()` for OpenAI compatibility
- Add conditional `Strict` mode to OpenAI `FunctionDefinition` when all properties are required

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `provider-openai-compatible`: Tool parameter schema conversion now reads `ParametersJsonSchema` (modern) before `Parameters` (legacy); strict mode conditionally enabled
- `tool-schema-builder`: Schema output now includes `additionalProperties: false`

## Impact

- `internal/adk/model.go` — `convertTools()` switch logic for schema extraction
- `internal/agent/schema.go` — `SchemaBuilder.Build()` adds `additionalProperties: false`
- `internal/adk/tools.go` — `buildInputSchema()` sets `AdditionalProperties` false schema
- `internal/provider/openai/openai.go` — `canUseStrictMode()` helper + `Strict` field on `FunctionDefinition`
