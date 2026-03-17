## Why

Three critical bugs in the automation systems (Cron, Background, Workflow) cause degraded user experience: cron jobs produce identical output every run due to isolated sessions lacking history context, job removal/pause fails to stop already-dispatched executions, and AI agents cannot resolve jobs by name. Background task cancellation has a race condition where `Cancelled` status gets overwritten by `Failed`. Workflow re-runs share session keys causing result contamination, and cancelled workflows continue executing pending steps.

## What Changes

- Inject recent execution history into cron job prompts so the LLM avoids repeating previous outputs
- Track in-flight cron job executions and cancel them on RemoveJob/PauseJob/Stop
- Add name-or-ID resolution for `cron_remove`, `cron_pause`, `cron_resume` tools
- Guard `Task.Fail()` and `Task.Complete()` to preserve `Cancelled` status in background tasks
- Add context cancellation check in background `execute()` after runner returns
- Include `runID` in workflow session keys to isolate re-runs
- Add context cancellation checks before step execution and after semaphore acquisition in workflow DAG runner

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `cron-scheduling`: Add history-aware prompt enrichment and in-flight execution cancellation
- `background-execution`: Fix cancel status race condition with terminal state guards
- `workflow-engine`: Include runID in session keys and add cancellation checks in DAG execution

## Impact

- `internal/cron/executor.go` — history injection via `buildPromptWithHistory()`
- `internal/cron/scheduler.go` — `inFlight` tracking, `cancelInFlight()`, `ResolveJobID()`
- `internal/app/tools_automation.go` — `cron_remove/pause/resume` name-or-ID resolution
- `internal/background/task.go` — `Fail()`/`Complete()` cancelled state guard
- `internal/background/manager.go` — context cancellation early return in `execute()`
- `internal/workflow/engine.go` — session key format change, cancellation checks
- No external API or dependency changes
