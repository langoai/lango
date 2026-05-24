## 1. ToolCapability.KnowledgeSaveable()

- [x] 1.1 Add `KnowledgeSaveable() bool` method on `ToolCapability` in `internal/agent/capability.go`
- [x] 1.2 Add unit tests: read-only true, activity read/query true, write/execute/manage/zero false

## 2. Catalog.SaveableToolNames()

- [x] 2.1 Add `SaveableToolNames() []string` method on `Catalog` in `internal/toolcatalog/catalog.go`
- [x] 2.2 Add unit test: register mixed tools, verify only saveable ones returned (sorted)

## 3. BuildHookRegistry catalog integration

- [x] 3.1 Add `catalog *toolcatalog.Catalog` parameter to `BuildHookRegistry`
- [x] 3.2 When catalog != nil, use `catalog.SaveableToolNames()` for KnowledgeSaveHook; when nil, fall back to `DefaultSaveableTools`
- [x] 3.3 Update private `buildHookRegistry` to pass catalog
- [x] 3.4 Update `cli/agent/hooks.go` to pass nil for catalog
- [x] 3.5 Add tests for both catalog and nil-catalog paths

## 4. CLI source indication

- [x] 4.1 In `cli/agent/hooks.go` KnowledgeSaveHook details, add `source: "fallback-constant"` field

## 5. Verification

- [x] 5.1 `go build ./...` passes
- [x] 5.2 `go test ./...` passes — zero FAIL
