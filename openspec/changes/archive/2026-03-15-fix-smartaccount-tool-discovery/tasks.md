## 1. Exec Guard

- [x] 1.1 Add `lango account` guard entry to `blockLangoExec` in `internal/app/tools.go`
- [x] 1.2 Add test case for `lango account deploy` in `internal/app/tools_test.go`

## 2. Disabled Category Registration

- [x] 2.1 Add `else` branch after `initSmartAccount()` in `internal/app/app.go` to register disabled `smartaccount` category
- [x] 2.2 Enhance `initSmartAccount()` log messages with config hint fields in `internal/app/wiring_smartaccount.go`

## 3. Diagnostic Tool

- [x] 3.1 Add `buildHealthTool()` function in `internal/toolcatalog/dispatcher.go`
- [x] 3.2 Update `BuildDispatcher()` to return 3 tools (list, invoke, health)
- [x] 3.3 Update `TestBuildDispatcher_ReturnsTwo` → `TestBuildDispatcher_ReturnsThree` in `dispatcher_test.go`
- [x] 3.4 Add `TestBuiltinHealth_ShowsDisabledCategories` test in `dispatcher_test.go`
- [x] 3.5 Update `TestDispatcher_SafetyLevels` to include health tool assertion

## 4. Orchestrator Prompt

- [x] 4.1 Add Diagnostics section to `buildOrchestratorInstruction()` in `internal/orchestration/tools.go`
- [x] 4.2 Update `TestBuildOrchestratorInstruction_DelegateOnly` to verify diagnostics section

## 5. Verification

- [x] 5.1 Run `go build ./...` and verify zero errors
- [x] 5.2 Run `go test ./internal/app/... ./internal/toolcatalog/... ./internal/orchestration/...` and verify all pass
