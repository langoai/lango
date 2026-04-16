## 1. ToolCapability.KnowledgeSaveable()

- [ ] 1.1 Add `KnowledgeSaveable() bool` method on `ToolCapability` in `internal/agent/capability.go` — returns true when `ReadOnly || Activity == ActivityRead || Activity == ActivityQuery`
- [ ] 1.2 Add unit tests for KnowledgeSaveable: read-only true, activity read true, activity query true, write false, default false

## 2. Catalog.SaveableToolNames()

- [ ] 2.1 Add `SaveableToolNames() []string` method on `Catalog` in `internal/toolcatalog/catalog.go` — filter registered tools by `KnowledgeSaveable()`, return sorted names
- [ ] 2.2 Add unit test: register mixed tools, verify only saveable ones returned

## 3. BuildHookRegistry catalog integration

- [ ] 3.1 Add `catalog *toolcatalog.Catalog` parameter to `BuildHookRegistry` in `internal/app/app.go`
- [ ] 3.2 When catalog != nil, use `catalog.SaveableToolNames()` for KnowledgeSaveHook; when nil, fall back to `DefaultSaveableTools`
- [ ] 3.3 Update private `buildHookRegistry` to pass `app.ToolCatalog`
- [ ] 3.4 Update `cli/agent/hooks.go` to pass nil for catalog
- [ ] 3.5 Update `BuildHookRegistry` tests for both catalog and nil-catalog paths

## 4. CLI source indication

- [ ] 4.1 In `cli/agent/hooks.go` KnowledgeSaveHook details, add `source: "fallback-constant"` field (since CLI always passes nil catalog)

## 5. Verification

- [ ] 5.1 `go build ./...` passes
- [ ] 5.2 `go test ./...` passes
