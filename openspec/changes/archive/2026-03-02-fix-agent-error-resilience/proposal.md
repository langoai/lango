## Why

Gemini API rejects requests with `INVALID_ARGUMENT` when session history contains consecutive same-role turns, missing FunctionResponse after FunctionCall, or non-user first turns. Additionally, agent runs exhaust the 25-turn limit prematurely because delegation events (agent-to-agent transfers) are counted as tool-calling turns. These are the two most frequent runtime errors during chat sessions.

## What Changes

- Add a 5-step content sanitization pipeline in the Gemini provider to enforce strict turn-ordering rules before every API call
- Add defense-in-depth consecutive role merging in EventsAdapter.All() to prevent malformed sequences from reaching any provider
- Exclude delegation events (TransferToAgent) from turn counting in agent.Run()
- Add graceful degradation: grant one wrap-up turn after limit reached before hard stop
- Add 80% turn limit warning log for observability
- Raise multi-agent default maxTurns from 25 to 50

## Capabilities

### New Capabilities
- `gemini-content-sanitization`: Gemini provider content turn-order sanitization pipeline and session event defense-in-depth merging

### Modified Capabilities
- `agent-turn-limit`: Delegation event exclusion, graceful wrap-up turn, 80% threshold warning, multi-agent default turn limit
- `multi-agent-orchestration`: Default turn limit raised to 50 when multiAgent mode is enabled and no explicit MaxTurns is configured

## Impact

- `internal/provider/gemini/sanitize.go` (NEW) — 5-step sanitization pipeline
- `internal/provider/gemini/sanitize_test.go` (NEW) — 10 table-driven tests
- `internal/provider/gemini/gemini.go` — call sanitizeContents before GenerateContentStream
- `internal/adk/agent.go` — turn counting rework with delegation exclusion, wrap-up turn, 80% warning
- `internal/adk/agent_test.go` — tests for hasFunctionCalls, isDelegationEvent
- `internal/adk/state.go` — consecutive role merging in EventsAdapter.All()
- `internal/adk/state_test.go` — updated tests + ConsecutiveRoleMerging tests
- `internal/app/wiring.go` — multi-agent default maxTurns = 50
