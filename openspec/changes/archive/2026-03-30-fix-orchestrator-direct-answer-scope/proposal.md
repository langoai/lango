## Why

The orchestrator LLM (gemini-3-flash-preview) classifies `weather` and `general knowledge` as direct-answer topics in its ASSESS step, but these requests require real-time data that the tool-less orchestrator cannot provide. When the LLM attempts to call tools directly (e.g., search, browser) to fulfill these requests, the direct-tool-call guard fires E003 (`orchestrator_direct_tool_call`) and recovery escalates immediately, leaving the user with a repeated `[E003] A tool execution failed` error on every attempt.

## What Changes

- Remove `weather` and `general knowledge` from the ASSESS step 0 direct-answer list in `buildOrchestratorInstruction()`
- Remove the same terms from the Delegation Rules #1 direct-answer list
- Add explicit "MUST NOT emit any function calls" guard to the ASSESS block to prevent tool hallucination even during direct response
- Narrow the direct-answer scope to `(greeting, opinion, math, small talk)` — topics that never require real-time data or tool access
- Add regression tests verifying the ASSESS and Delegation Rules strings exclude removed terms

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `multi-agent-orchestration`: Change the orchestrator's direct-answer classification criteria — reclassify `weather`/`general knowledge` as delegation targets, add function-call prohibition guard to ASSESS step

## Impact

- `internal/orchestration/tools.go` — prompt string changes in `buildOrchestratorInstruction()`
- `internal/orchestration/orchestrator_test.go` — new regression tests for ASSESS and Delegation Rules content
- No changes to recovery policy (`internal/agentrt/recovery.go`) — RecoveryEscalate preserved per spec
