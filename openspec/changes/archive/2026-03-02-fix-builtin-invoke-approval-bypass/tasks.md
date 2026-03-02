## 1. Dispatcher Safety Check

- [x] 1.1 Add safety level check in `builtin_invoke` handler (`internal/toolcatalog/dispatcher.go`) — block tools with SafetyLevel >= Dangerous, return error directing to sub-agent delegation
- [x] 1.2 Update `TestBuiltinInvoke_Success` to use a safe tool (browser_navigate) instead of dangerous tool (exec_shell)
- [x] 1.3 Add `TestBuiltinInvoke_BlocksDangerousTools` test case verifying dangerous tools are rejected with "requires approval" error

## 2. Orchestrator Tool Removal

- [x] 2.1 Remove `universalTools` construction and `UniversalTools` assignment from orchestrator config in `internal/app/wiring.go`
- [x] 2.2 Restore orchestrator identity prompt to always say "delegate to sub-agents instead" (remove catalog-conditional branch in `wiring.go`)
- [x] 2.3 Remove universal tool adaptation block from `BuildAgentTree` in `internal/orchestration/orchestrator.go` — orchestrator gets no tools
- [x] 2.4 Remove `hasUniversalTools` parameter from `buildOrchestratorInstruction` in `internal/orchestration/tools.go` — always emit delegation-only prompt

## 3. Test Updates

- [x] 3.1 Update all `buildOrchestratorInstruction` test calls in `orchestrator_test.go` to match new signature (remove `hasUniversalTools` argument)
- [x] 3.2 Replace `TestBuildOrchestratorInstruction_WithUniversalTools` and `WithoutUniversalTools` tests with single `TestBuildOrchestratorInstruction_DelegateOnly` test
- [x] 3.3 Verify `go build ./...` passes
- [x] 3.4 Verify `go test ./internal/toolcatalog/... ./internal/orchestration/...` passes
