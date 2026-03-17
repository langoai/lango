## Why

When `agent.multiAgent=true`, cron jobs, background tasks, and workflow steps fail because the orchestrator's Decision Protocol Step 0 misclassifies their prompts as "simple conversational requests" and attempts to respond directly without tools. The root cause is that automation prompts lack a signal distinguishing them from user chat, and the orchestrator has no rule to override Step 0 for automated tasks.

## What Changes

- Add `[Automated Task]` prefix to all prompts emitted by cron executor, background task manager, and workflow engine before they reach the agent runner.
- Add an "Automated Task Handling" section to the orchestrator instruction that overrides the ASSESS step for prefixed prompts, ensuring delegation to the correct sub-agent based on task content.

## Capabilities

### New Capabilities
- `automation-prompt-routing`: Enriches automation system prompts with a machine-readable prefix and adds orchestrator routing rules so automated tasks are always delegated to sub-agents.

### Modified Capabilities

## Impact

- `internal/cron/executor.go` — `buildPromptWithHistory()` prepends automation prefix
- `internal/background/manager.go` — `execute()` wraps prompt with automation prefix
- `internal/workflow/engine.go` — `executeStep()` wraps rendered prompt with automation prefix
- `internal/orchestration/tools.go` — `buildOrchestratorInstruction()` gains "Automated Task Handling" section
- Existing tests updated to expect the new prefix in prompts
