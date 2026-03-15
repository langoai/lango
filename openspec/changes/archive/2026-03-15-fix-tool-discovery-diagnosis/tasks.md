## 1. Catalog Query Methods

- [x] 1.1 Add `ToolNamesForCategory(category string) []string` to Catalog
- [x] 1.2 Add `EnabledCategorySummary() map[string][]string` to Catalog

## 2. Enhanced builtin_health

- [x] 2.1 Update builtin_health to include tool name lists per enabled category
- [x] 2.2 Update builtin_health disabled hint to use actionable `lango config set` format
- [x] 2.3 Add test for tool names in builtin_health output

## 3. Disabled Category Registration

- [x] 3.1 Register disabled cron category when cron.enabled is false
- [x] 3.2 Register disabled background category when background.enabled is false
- [x] 3.3 Register disabled workflow category when workflow.enabled is false

## 4. Diagnostic Logging

- [x] 4.1 Add `logToolRegistrationSummary()` function to app.go
- [x] 4.2 Call diagnostic log after all tool registration is complete

## 5. System Prompt Tool Catalog Section

- [x] 5.1 Add `SectionToolCatalog` to prompt section IDs
- [x] 5.2 Implement `buildToolCatalogSection()` in wiring.go
- [x] 5.3 Wire tool catalog section into initAgent prompt builder

## 6. Orchestrator Routing Enhancement

- [x] 6.1 Add `ToolNames` field to `routingEntry` struct
- [x] 6.2 Populate `ToolNames` in `buildRoutingEntry()`
- [x] 6.3 Render tool names in orchestrator instruction
- [x] 6.4 Update orchestrator test to pass tools to buildRoutingEntry

## 7. Verification

- [x] 7.1 Verify `go build ./...` passes
- [x] 7.2 Verify `go test ./internal/toolcatalog/...` passes
- [x] 7.3 Verify `go test ./internal/prompt/...` passes
- [x] 7.4 Verify `go test ./internal/orchestration/...` passes
