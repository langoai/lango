## Context

`buildInputSchema()` in `internal/adk/tools.go` converts `agent.Tool.Parameters` (a `map[string]interface{}`) into a `*jsonschema.Schema` for ADK tool registration. It was written to handle two flat formats: `ParameterDef` structs and raw `map[string]interface{}` per property. However, `SchemaBuilder.Build()` returns a full JSON Schema object where top-level keys are `type`, `properties`, and `required` — not parameter names. The function iterates these keys as if they were parameters, producing a broken schema.

## Goals / Non-Goals

**Goals:**
- Fix `buildInputSchema()` to correctly detect and parse full JSON Schema format from `SchemaBuilder.Build()`
- Preserve backward compatibility with existing flat ParameterDef and flat map formats
- Add enum extraction for map-based property definitions

**Non-Goals:**
- Changing `SchemaBuilder.Build()` output format
- Modifying any tool definition files (they are already correct)
- Supporting deeply nested JSON Schema features (e.g., `$ref`, `oneOf`, `allOf`)

## Decisions

**Detection via `"properties"` key**: Check if `params["properties"]` exists and is a `map[string]interface{}`. If so, treat the input as full JSON Schema format. This is unambiguous because no tool parameter would be named `"properties"` with a map value containing property definitions.

**Alternative considered**: Type-switch on a wrapper struct — rejected because `Build()` returns `map[string]interface{}` and changing that would break all callers.

**`params` reassignment**: After detecting full schema format, reassign `params` to the nested properties map. This allows the existing per-property iteration logic (ParameterDef, map, fallback) to work unchanged on the inner property definitions.

**`required` extraction before loop**: Extract the top-level `required` array before entering the property loop. Handle both `[]string` (from Go code) and `[]interface{}` (from JSON deserialization) representations.

## Risks / Trade-offs

**[Risk] Flat map with a key literally named "properties"** → Extremely unlikely in practice; no existing tool uses this. The detection also requires the value to be a `map[string]interface{}`, further reducing false positives.

**[Risk] SchemaBuilder enum values stored as `[]string` not `[]interface{}`** → Added `[]string` to `[]any` conversion in the map branch alongside existing `[]interface{}` handling.
