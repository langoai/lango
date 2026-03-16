## 1. Fix buildInputSchema

- [x] 1.1 Add full JSON Schema detection in `buildInputSchema()` — check for `"properties"` key containing a map, extract nested properties and top-level `required` array
- [x] 1.2 Add enum extraction for map-based property definitions (`[]string` and `[]interface{}`)

## 2. Tests

- [x] 2.1 Add `TestBuildInputSchema_SchemaBuilder` — table-driven test for SchemaBuilder.Build() output with various param combinations
- [x] 2.2 Add `TestBuildInputSchema_SchemaBuilder_PropertyTypes` — verify type, description, and enum extraction for all property types
- [x] 2.3 Add `TestBuildInputSchema_FlatParameterDef` — regression test for existing ParameterDef format
- [x] 2.4 Add `TestBuildInputSchema_FlatMap` — regression test for existing flat map format
- [x] 2.5 Add `TestAdaptTool_SchemaBuilder` — end-to-end AdaptTool with SchemaBuilder parameters

## 3. Verification

- [x] 3.1 Run `go build ./...` — verify no build errors
- [x] 3.2 Run `go test ./internal/adk/...` — verify all tests pass
