## 1. Critical Wiring / Lifecycle Fixes

- [x] 1.1 Add `Shutdown()` method to BackgroundManager that cancels all Pending/Running tasks
- [x] 1.2 Add `Shutdown()` method to WorkflowEngine that cancels all running workflows via cancels map
- [x] 1.3 Add BackgroundManager.Shutdown() and WorkflowEngine.Shutdown() calls to App.Stop()
- [x] 1.4 Fix Workflow CLI initEngine nil logger — use zap.NewProduction() fallback
- [x] 1.5 Add `StartedAt time.Time` field to RunStatus struct in workflow/step.go
- [x] 1.6 Populate StartedAt in ListRuns() and GetRunStatus() from workflow run record
- [x] 1.7 Replace time.Now() with RunStatus.StartedAt in workflow CLI list command

## 2. Agent Tool Implementation

- [x] 2.1 Add `History(ctx, jobID, limit)` and `AllHistory(ctx, limit)` delegation methods to cron Scheduler
- [x] 2.2 Implement `buildCronTools()` with 6 tools: cron_add, cron_list, cron_pause, cron_resume, cron_remove, cron_history
- [x] 2.3 Implement `buildBackgroundTools()` with 5 tools: bg_submit, bg_status, bg_list, bg_result, bg_cancel
- [x] 2.4 Implement `buildWorkflowTools()` with 5 tools: workflow_run, workflow_status, workflow_list, workflow_cancel, workflow_save
- [x] 2.5 Update `buildApprovalSummary()` with cases for cron_add, cron_remove, bg_submit, workflow_run, workflow_cancel

## 3. App Initialization Reorder & Tool Wiring

- [x] 3.1 Move cron/bg/workflow initialization from after channels to before approval wrapping (steps 5j-5l)
- [x] 3.2 Register cron/bg/workflow tools via buildXxxTools() into the tools slice before approval wrapping
- [x] 3.3 Verify agentRunnerAdapter lazy pattern still works with early initialization

## 4. Multi-Agent Routing

- [x] 4.1 Add "automator" AgentSpec with cron_/bg_/workflow_ prefixes and automation keywords
- [x] 4.2 Add `Automator []*agent.Tool` field to RoleToolSet
- [x] 4.3 Add automator case to PartitionTools (before operator matching)
- [x] 4.4 Add automator case to toolsForSpec
- [x] 4.5 Add cron_, bg_, workflow_ entries to capabilityMap

## 5. System Prompt Enhancement

- [x] 5.1 Add `SectionAutomation SectionID = "automation"` constant to prompt/section.go
- [x] 5.2 Implement `buildAutomationPromptSection()` in wiring.go with dynamic capability listing
- [x] 5.3 Wire automation prompt section in initAgent() when any automation system is enabled

## 6. Notification Enhancements

- [x] 6.1 Add `DeliverStart(ctx, jobName, targets)` method to cron Delivery
- [x] 6.2 Call DeliverStart before runner.Run() in cron executor Execute()
- [x] 6.3 Change nil sender log level from Debug to Warn in cron Delivery.Deliver()
- [x] 6.4 Add `NotifyStart(ctx, task)` method to background Notification
- [x] 6.5 Call NotifyStart after task.SetRunning() in background manager execute()

## 7. Verification

- [x] 7.1 Run `go build ./...` — full build passes
- [x] 7.2 Run `go test ./...` — all tests pass
