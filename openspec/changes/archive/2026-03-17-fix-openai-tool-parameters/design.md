## Context

`genai.FunctionDeclaration` has two schema fields:
- `Parameters *Schema` (legacy, Gemini native)
- `ParametersJsonSchema any` (modern, ADK v0.5.0+)

ADK's `functiontool.Declaration()` sets `ParametersJsonSchema` and leaves `Parameters` nil. Our `convertTools()` only reads `Parameters`, so OpenAI receives `"parameters": {}` for every tool.

Additionally, OpenAI produces better tool calls when `additionalProperties: false` is present, and supports a `Strict` mode that constrains output to the exact schema.

## Goals / Non-Goals

**Goals:**
- Fix tool parameter schema extraction so OpenAI receives complete parameter definitions
- Add `additionalProperties: false` to all tool schemas for improved accuracy
- Enable OpenAI `Strict` mode for tools where all properties are required

**Non-Goals:**
- Changing how legacy Gemini-native tools work (backward compatibility preserved)
- Refactoring the broader tool adaptation pipeline
- Supporting nested object schemas or `$ref` in strict mode validation

## Decisions

### 1. Switch statement for schema extraction priority

Use a `switch` statement in `convertTools()` that checks `ParametersJsonSchema` first, then falls back to `Parameters`. This preserves backward compatibility while supporting the modern ADK path.

**Alternative considered**: Always converting `ParametersJsonSchema` only — rejected because legacy tools still use `Parameters`.

### 2. `{Not: {}}` pattern for false schema

The `jsonschema.Schema` type has no boolean representation. The `{Not: {}}` pattern (`"not": {}`) serializes to `"additionalProperties": false` in JSON, which is the standard JSON Schema "false schema" idiom.

**Alternative considered**: Custom JSON marshaler — rejected as unnecessary complexity when the library handles this natively.

### 3. Conditional strict mode

OpenAI's strict mode requires all properties to be in `required` and `additionalProperties: false`. Tools with optional parameters cannot use strict mode (API rejects them). A `canUseStrictMode()` helper inspects the schema at conversion time.

**Alternative considered**: Always enabling strict mode — rejected because it would break tools with optional parameters.

## Risks / Trade-offs

- [Schema format mismatch] → `json.Marshal`/`Unmarshal` roundtrip normalizes both schema types to `map[string]interface{}`, which may lose type information. Mitigated by the fact that this is the same serialization path used in the provider layer.
- [Strict mode false negatives] → `canUseStrictMode()` checks `required` as both `[]string` and `[]interface{}` since JSON unmarshaling may produce either type. Covered by tests.
