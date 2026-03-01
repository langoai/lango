## 1. AllowImport Guard

- [x] 1.1 Add AllowImport check at the start of import_skill handler in tools_meta.go
- [x] 1.2 Verify import_skill returns error when AllowImport is false

## 2. Tool Catalog Package

- [x] 2.1 Create internal/toolcatalog/catalog.go with Catalog, Category, ToolEntry, ToolSchema types
- [x] 2.2 Implement Register, Get, ListCategories, ListTools, ToolCount methods
- [x] 2.3 Create internal/toolcatalog/dispatcher.go with BuildDispatcher returning builtin_list and builtin_invoke
- [x] 2.4 Create internal/toolcatalog/catalog_test.go with table-driven tests for catalog operations
- [x] 2.5 Create internal/toolcatalog/dispatcher_test.go with tests for list, invoke, not-found, safety levels

## 3. App Wiring

- [x] 3.1 Add ToolCatalog field to App struct in types.go
- [x] 3.2 Create catalog and register categories/tools in app.go for all tool builders
- [x] 3.3 Add dispatcher tools to the tools slice before approval wrapping
- [x] 3.4 Update initAgent signature to accept catalog parameter in wiring.go
- [x] 3.5 Wire UniversalTools into orchestration.Config in wiring.go

## 4. Multi-Agent Orchestration

- [x] 4.1 Add UniversalTools field to orchestration.Config
- [x] 4.2 Adapt and assign universal tools to orchestrator in BuildAgentTree
- [x] 4.3 Update buildOrchestratorInstruction to accept hasUniversalTools parameter
- [x] 4.4 Add builtin_ prefix skip in PartitionTools
- [x] 4.5 Update orchestrator identity prompt in wiring.go for universal tool awareness

## 5. Message Updates

- [x] 5.1 Update blockLangoExec catch-all message with builtin_list hint in tools.go

## 6. Tests

- [x] 6.1 Add TestPartitionTools_SkipsBuiltinPrefix in orchestrator_test.go
- [x] 6.2 Add TestBuildOrchestratorInstruction_WithUniversalTools in orchestrator_test.go
- [x] 6.3 Add TestBuildOrchestratorInstruction_WithoutUniversalTools in orchestrator_test.go
- [x] 6.4 Update existing buildOrchestratorInstruction test calls for new signature

## 7. Verification

- [x] 7.1 Run go build ./... with no errors
- [x] 7.2 Run go test ./... with all tests passing
