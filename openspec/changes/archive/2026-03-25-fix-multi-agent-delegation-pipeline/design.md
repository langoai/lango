## Overview

Three independent bugs in the multi-agent structured orchestration pipeline prevented
vault (and other isolated sub-agent) delegation from completing. Each bug was confirmed
via DB turntrace evidence and fixed with a targeted, minimal change.

## Architecture

No new components. All changes are surgical fixes within the existing execution pipeline:

```
User → TurnRunner → CoordinatingExecutor → ADK Agent.Run() → provider API
                                              │
                        Fix A: guard exception │ (agent.go:331)
                        Fix B: FuncResp split  │ (model.go:convertMessages)
                        Fix C: author preserve │ (session_service.go:closeDangling)
                        Fix D: recovery escape │ (recovery.go:Decide)
                        Fix E: TUI log redirect│ (main.go:runChat)
```

## Design Decisions

### 1. isPureTransferToAgentCall guard (agent.go)

ADK yields the model-response event (with transfer_to_agent FunctionCall) BEFORE
promoting it to Actions.TransferToAgent in a subsequent event. The guard must allow
pure transfer_to_agent FunctionCalls through while still blocking real tool calls
from the orchestrator.

Decision: `isPureTransferToAgentCall` checks ALL FunctionCalls are transfer_to_agent.
Mixed events (transfer + real tool) still trigger the guard.

### 2. convertMessages FunctionResponse split (model.go)

EventsAdapter.All() merges consecutive same-role events (state.go:301-304). When
vault calls 3 tools in parallel, 3 tool-role events merge into 1 Content with 3
FunctionResponse parts. convertMessages then overwrites tool_call_id metadata —
only the last ID survives.

Decision: Split at convertMessages level (not at EventsAdapter merge level) to
minimize blast radius. Only trigger when role=="tool" AND FunctionResponse count >= 2.

### 3. OriginAuthor in closeDanglingParentToolCalls (session_service.go)

danglingToolCalls() now tracks which assistant message emitted each tool call,
preserving the originating agent's Author in synthetic closure messages.

Fallback: rootAgentName → "lango-agent". Warning logged on empty OriginAuthor.

### 4. CauseOrchestratorDirectTool → RecoveryEscalate (recovery.go)

Same-input retry cannot fix a guard violation. Changed from RecoveryRetry to
RecoveryEscalate to prevent the 3-iteration retry loop observed in traces.

### 5. stdlib logger redirect (main.go)

ADK uses Go stdlib log.Printf (not zap). Added log.SetOutput(logFile) after
logging.Init() with defer logFile.Close() for cleanup.

## Non-Goals

- EventsAdapter merge logic change (state.go:301-304) — deferred, Fix B handles it at conversion
- EventsAdapter tool-role Author fallback (state.go:272) — deferred until blank-author evidence found
- Gemini thought_signature fix — high probability of same merge root cause, but unconfirmed
