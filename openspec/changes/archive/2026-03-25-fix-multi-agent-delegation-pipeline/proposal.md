## Why

Vault handoff fails in multi-agent structured orchestration mode due to 3 distinct bugs.
(1) Timing conflict between ADK 2-phase delegation and Lango guard, (2) ID loss in parallel tool responses,
(3) Auxiliary issues (TUI log leak, dangling cleanup author loss, recovery infinite loop).
Each root cause was confirmed through DB/turntrace evidence.

## What Changes

- Add exception for `transfer_to_agent` FunctionCall to pass through orchestrator direct-tool guard (ADK 2-phase compatibility)
- Split multiple FunctionResponses merged by EventsAdapter in `convertMessages` into individual provider.Messages
- Preserve originating agent Author in `closeDanglingParentToolCalls` synthetic tool response
- Change `CauseOrchestratorDirectTool` recovery to Escalate (prevent same-input retry infinite loop)
- Redirect Go stdlib logger to log file in TUI (block ADK's log.Printf leak)
- Improve `repairOrphanedToolCalls` error messages to be more accurate and actionable
- Add error classification diagnostic logging to CoordinatingExecutor recovery

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `agent-control-plane`: Added transfer_to_agent exception to orchestrator direct-tool guard, added CauseOrchestratorDirectTool escalation to recovery
- `agent-error-handling`: Improved repairOrphanedToolCalls message, preserved dangling cleanup author, added convertMessages FunctionResponse split

## Impact

- `internal/adk/agent.go` — isPureTransferToAgentCall guard exception
- `internal/adk/model.go` — convertMessages multi-FunctionResponse split
- `internal/adk/session_service.go` — danglingCall OriginAuthor, diagnostic logging
- `internal/agentrt/coordinating_executor.go` — recovery diagnostic logging
- `internal/agentrt/recovery.go` — CauseOrchestratorDirectTool escalation
- `internal/provider/openai/openai.go` — improved orphan error message
- `cmd/lango/main.go` — stdlib logger redirect
