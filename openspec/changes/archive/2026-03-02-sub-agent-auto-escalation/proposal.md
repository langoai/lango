## Why

Sub-agents that receive out-of-scope requests fail silently — they attempt to answer, produce unhelpful text like "[REJECT] ask another agent", or hallucinate non-existent agent names. The user must manually re-route, which is a critical UX failure in multi-agent orchestration. The root cause is that the `[REJECT]` protocol exists only as a text convention with no code-level enforcement, and sub-agents lack instructions to use the ADK `transfer_to_agent` tool for escalation.

## What Changes

- Replace `[REJECT]` text protocol in all 7 sub-agent instructions with `transfer_to_agent` call to `lango-orchestrator` (Escalation Protocol)
- Add Step 0 (ASSESS) to orchestrator's Decision Protocol so it handles simple conversational requests directly
- Replace orchestrator's "Rejection Handling" section with "Re-Routing Protocol" for when sub-agents transfer back
- Reorder Delegation Rules to prioritize direct response over delegation
- Add `[REJECT]` text detection safety net in `RunAndCollect` to auto-retry when a sub-agent emits reject text instead of using the tool

## Capabilities

### New Capabilities

(none — this enhances existing capabilities)

### Modified Capabilities

- `multi-agent-orchestration`: Sub-agent escalation changes from text-based `[REJECT]` protocol to `transfer_to_agent` tool calls; orchestrator gains re-routing protocol and direct-response assessment step
- `agent-self-correction`: `RunAndCollect` gains `[REJECT]` text detection as a safety net, auto-retrying with re-routing instruction when detected

## Impact

- `internal/orchestration/tools.go` — All 7 sub-agent instruction prompts and orchestrator prompt builder
- `internal/adk/agent.go` — `RunAndCollect` method with REJECT detection/retry logic
- `internal/orchestration/orchestrator_test.go` — Updated assertions for new protocol
- `internal/adk/agent_test.go` — New unit tests for REJECT pattern detection
