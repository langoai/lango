## Why

Cron, Background, and Workflow systems are implemented but only accessible via CLI. Users cannot schedule tasks, submit background jobs, or run workflows through natural conversation (e.g., "Schedule a news summary every morning at 9am"). Additionally, these systems lack proper shutdown lifecycle methods and start notifications.

## What Changes

- Add 16 agent tools: 6 cron (`cron_add`, `cron_list`, `cron_pause`, `cron_resume`, `cron_remove`, `cron_history`), 5 background (`bg_submit`, `bg_status`, `bg_list`, `bg_result`, `bg_cancel`), 5 workflow (`workflow_run`, `workflow_status`, `workflow_list`, `workflow_cancel`, `workflow_save`)
- Add `Shutdown()` methods to `BackgroundManager` and `WorkflowEngine` for graceful lifecycle management
- Add `History()`/`AllHistory()` delegation methods to `cron.Scheduler`
- Add start notifications (`DeliverStart`, `NotifyStart`) for cron jobs and background tasks
- Reorder `App.New()` initialization so automation tools are registered before approval wrapping
- Add `BackgroundManager.Shutdown()` and `WorkflowEngine.Shutdown()` calls in `App.Stop()`
- Add "automator" agent spec for multi-agent routing of automation tools
- Add `SectionAutomation` prompt section to guide the agent on automation tool usage
- Fix `time.Now()` TODO in workflow CLI list command
- Fix nil logger in workflow CLI `initEngine()`
- Change cron delivery nil-sender log level from Debug to Warn

## Capabilities

### New Capabilities
- `automation-agent-tools`: Agent-facing tools for cron scheduling, background task management, and workflow execution via conversational interface

### Modified Capabilities
- `cron-scheduling`: Added `History()`/`AllHistory()` scheduler delegation, `DeliverStart()` notification, Debug→Warn log level for nil sender
- `background-execution`: Added `Shutdown()` lifecycle method, `NotifyStart()` notification
- `workflow-engine`: Added `Shutdown()` lifecycle method, `StartedAt` field in `RunStatus`
- `multi-agent-orchestration`: Added "automator" agent spec and routing for `cron_`, `bg_`, `workflow_` prefixed tools
- `structured-prompt-builder`: Added `SectionAutomation` constant for automation prompt section
- `bootstrap-lifecycle`: Reordered `App.New()` to init cron/bg/workflow before agent; added shutdown calls in `App.Stop()`

## Impact

- **Files modified**: `internal/app/app.go`, `internal/app/tools.go`, `internal/app/wiring.go`, `internal/background/manager.go`, `internal/background/notification.go`, `internal/workflow/engine.go`, `internal/workflow/step.go`, `internal/workflow/state.go`, `internal/cron/scheduler.go`, `internal/cron/executor.go`, `internal/cron/delivery.go`, `internal/orchestration/tools.go`, `internal/prompt/section.go`, `internal/cli/workflow/workflow.go`
- **No breaking changes**: All changes are additive; existing CLI commands and APIs remain unchanged
- **No new dependencies**: Uses existing packages only
