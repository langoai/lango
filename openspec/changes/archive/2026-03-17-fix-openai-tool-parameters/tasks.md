## 1. Schema Extraction Fix (P0)

- [x] 1.1 Update `convertTools()` in `internal/adk/model.go` to check `ParametersJsonSchema` before `Parameters`
- [x] 1.2 Add tests: `TestConvertTools_ParametersJsonSchema`, `TestConvertTools_LegacyParameters`, `TestConvertTools_BothSet_ParametersJsonSchemaPriority`

## 2. Additional Properties (P1)

- [x] 2.1 Add `"additionalProperties": false` to `SchemaBuilder.Build()` in `internal/agent/schema.go`
- [x] 2.2 Set `AdditionalProperties` false schema (`{Not: {}}`) in `buildInputSchema()` in `internal/adk/tools.go`
- [x] 2.3 Add test assertion for `additionalProperties` in `schema_test.go`
- [x] 2.4 Add `TestBuildInputSchema_AdditionalPropertiesFalse` in `tools_test.go`

## 3. OpenAI Strict Mode (P2)

- [x] 3.1 Implement `canUseStrictMode()` helper in `internal/provider/openai/openai.go`
- [x] 3.2 Apply `Strict` field conditionally in `convertParams()`
- [x] 3.3 Add `TestCanUseStrictMode` table-driven tests
- [x] 3.4 Add `TestConvertParams_StrictMode` integration test

## 4. Verification

- [x] 4.1 `go build ./...` passes
- [x] 4.2 `go test ./internal/adk/ ./internal/agent/ ./internal/provider/openai/` passes
- [x] 4.3 `go test ./...` passes
