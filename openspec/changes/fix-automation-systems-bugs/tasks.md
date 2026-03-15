## 1. Cron — History-aware prompt enrichment

- [x] 1.1 Add `recentHistoryLimit` and `maxResultPreviewLen` constants to `internal/cron/executor.go`
- [x] 1.2 Implement `buildPromptWithHistory(ctx, job)` method on Executor that queries history and prepends previous outputs
- [x] 1.3 Update `Execute()` to call `buildPromptWithHistory()` and pass enriched prompt to runner (save original prompt in history)
- [x] 1.4 Add test `TestExecutor_Execute_InjectsHistoryContext` — history entries present → enriched prompt sent to runner
- [x] 1.5 Add test `TestExecutor_Execute_NoHistory_OriginalPrompt` — no history → original prompt unchanged
- [x] 1.6 Add test `TestExecutor_Execute_HistoryQueryError_Graceful` — query error → fallback to original prompt

## 2. Cron — In-flight execution cancellation

- [x] 2.1 Add `inFlight map[string]context.CancelFunc` and `inFlightMu sync.Mutex` to Scheduler struct
- [x] 2.2 Initialize `inFlight` map in `New()` constructor
- [x] 2.3 Register cancel func in `executeWithSemaphore()` and defer cleanup
- [x] 2.4 Add `cancelInFlight(id)` helper method
- [x] 2.5 Call `cancelInFlight()` in `RemoveJob()` and `PauseJob()`
- [x] 2.6 Cancel all in-flight executions in `Stop()` before cron runner drain
- [x] 2.7 Add test `TestScheduler_RemoveJob_CancelsInFlight`

## 3. Cron — Name-or-ID job resolution

- [x] 3.1 Add `ResolveJobID(ctx, nameOrID)` method to Scheduler
- [x] 3.2 Update `cron_remove` handler in `tools_automation.go` to use `ResolveJobID`
- [x] 3.3 Update `cron_pause` handler to use `ResolveJobID`
- [x] 3.4 Update `cron_resume` handler to use `ResolveJobID`
- [x] 3.5 Update tool parameter descriptions to say "The cron job ID or name"
- [x] 3.6 Add tests `TestScheduler_ResolveJobID_ByUUID`, `ByName`, `NotFound`

## 4. Background — Cancel status guard

- [x] 4.1 Add `Cancelled` status guard in `Task.Fail()` — skip if already cancelled
- [x] 4.2 Add `Cancelled` status guard in `Task.Complete()` — skip if already cancelled
- [x] 4.3 Add context cancellation early return in `Manager.execute()` after runner returns
- [x] 4.4 Add test `TestTask_Fail_PreservesCancelledStatus`
- [x] 4.5 Add test `TestTask_Complete_PreservesCancelledStatus`
- [x] 4.6 Add test `TestManager_Cancel_PreservesStatus` — end-to-end cancel during execution

## 5. Workflow — Session key isolation and cancellation

- [x] 5.1 Change session key format in `executeStep()` to include runID: `workflow:{name}:{runID}:{stepID}`
- [x] 5.2 Add context cancellation check at start of `executeStep()`
- [x] 5.3 Add context cancellation check in `runDAG()` goroutine after semaphore acquisition
- [x] 5.4 Add test `TestEngine_SessionKeyFormat` — verify runID inclusion in key format
- [x] 5.5 Add test `TestEngine_ExecuteStep_ChecksCancellation` — pre-cancelled context returns immediately
- [x] 5.6 Add test `TestEngine_ExecuteStep_RunnerError` — error propagation with cancelled context

## 6. Verification

- [x] 6.1 Run `go build ./...` — full project build
- [x] 6.2 Run `go test ./internal/cron/...` — all cron tests pass
- [x] 6.3 Run `go test ./internal/background/...` — all background tests pass
- [x] 6.4 Run `go test ./internal/workflow/...` — all workflow tests pass
- [x] 6.5 Run `go test ./internal/app/...` — app integration tests pass
- [x] 6.6 Add `listHistoryErr` field to mock store for history query error testing
