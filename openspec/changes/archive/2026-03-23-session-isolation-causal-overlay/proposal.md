## Why

`SessionIsolation=true` currently routes isolated sub-agent events into child history only. That preserves cross-turn isolation, but it breaks ADK's same-run assumption that `AppendEvent` writes to the same causal event stream later read by `Events()`. As a result, isolated tool-bearing agents cannot see their own `FunctionResponse` and can loop on the same tool call.

## What Changes

- Redefine `SessionIsolation` as cross-turn persistence isolation, not same-run visibility isolation.
- Expose isolated child events to the parent session's in-memory view for the current run only.
- Remove raw overlay events before merge/discard, then write only a root-authored summary or compact failure note to the persistent parent history.
- Keep child raw history runtime-scoped and in-memory only.

## Capabilities

### Modified Capabilities

- `sub-session-isolation`: child sessions preserve persistent isolation while allowing same-run causal visibility.
- `multi-agent-orchestration`: isolated sub-agents can continue ADK tool loops without reissuing already-completed tool calls.

## Impact

- `internal/adk/session_service.go`
- `internal/adk/agent.go`
- `internal/adk/session_service_test.go`
- `openspec/specs/sub-session-isolation/spec.md`
- `openspec/specs/multi-agent-orchestration/spec.md`
- `docs/architecture/session-isolation-conflict.md`
- `docs/features/multi-agent.md`
- `docs/features/agent-format.md`
- `README.md`
