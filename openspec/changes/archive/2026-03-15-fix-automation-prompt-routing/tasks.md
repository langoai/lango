## 1. Automation Prompt Prefix

- [x] 1.1 Add `automationPrefix` constant and prepend it in `buildPromptWithHistory()` in `internal/cron/executor.go`
- [x] 1.2 Add `automationPrefix` constant and wrap prompt in `execute()` in `internal/background/manager.go`
- [x] 1.3 Add `automationPrefix` constant and wrap rendered prompt in `executeStep()` in `internal/workflow/engine.go`

## 2. Orchestrator Routing Rule

- [x] 2.1 Add "Automated Task Handling" section to `buildOrchestratorInstruction()` in `internal/orchestration/tools.go`, placed before Decision Protocol

## 3. Tests

- [x] 3.1 Update cron executor tests to assert `[Automated Task` prefix and `Task:` label in prompts
- [x] 3.2 Verify background manager tests pass with enriched prompt
- [x] 3.3 Verify workflow engine tests pass with enriched prompt
- [x] 3.4 Add `TestBuildOrchestratorInstruction_HasAutomatedTaskHandling` to orchestration tests

## 4. Verification

- [x] 4.1 Run `go build ./...` — confirm clean build
- [x] 4.2 Run `go test ./internal/cron/... ./internal/background/... ./internal/workflow/... ./internal/orchestration/...` — all pass
