## 1. Orchestrator Prompt Fix

- [x] 1.1 Remove `weather` and `general knowledge` from ASSESS step 0 direct-answer list in `buildOrchestratorInstruction()` (`internal/orchestration/tools.go:663`)
- [x] 1.2 Add "MUST NOT emit any function calls" guard to ASSESS block, with delegation instruction for real-time data requests
- [x] 1.3 Remove `weather` and `general knowledge` from Delegation Rules #1 (`internal/orchestration/tools.go:712`)

## 2. Regression Tests

- [x] 2.1 Update `TestBuildOrchestratorInstruction_HasAssessStep` to verify ASSESS direct-answer list excludes `weather` and `general knowledge`, and includes "MUST NOT emit any function calls" guard
- [x] 2.2 Add `TestBuildOrchestratorInstruction_DelegationRulesNoWeather` to verify Delegation Rules section excludes `weather` and `general knowledge`

## 3. Verification

- [x] 3.1 `CGO_ENABLED=1 go build -tags fts5 ./...` passes
- [x] 3.2 `CGO_ENABLED=1 go test -tags fts5 ./...` passes (all packages, including new tests)
- [ ] 3.3 Manual smoke check (when live Gemini env available): TUI request "tell me today's weather" does not produce E003 / `orchestrator_direct_tool_call`
