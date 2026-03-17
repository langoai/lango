## Why

The current fixed 5-minute request timeout causes failures for complex queries (multi-tool-call, long reasoning) that are still actively processing. Additionally, when a timeout occurs, incomplete session history leaks into the next turn, causing garbled responses. An idle-based timeout approach allows active requests to continue while still catching stuck/stalled agents.

## What Changes

- Extract `ExtendableDeadline` from `internal/app/` to a shared `internal/deadline/` package with `Reason()` tracking (idle vs max_timeout vs cancelled)
- Add `IdleTimeout` config field to `AgentConfig` — when set, the request stays alive as long as the agent produces activity (streaming chunks, tool calls), timing out only after inactivity
- Apply idle timeout to both channel handlers (`channels.go`) and gateway (`server.go`)
- Add `AnnotateTimeout` to the session `Store` interface — on timeout, append a synthetic assistant message to close the conversation turn cleanly
- Add `ErrIdleTimeout` (E006) error code for idle-specific timeout classification
- Backward-compatible: existing configs with only `requestTimeout` or `autoExtendTimeout` behave identically

## Capabilities

### New Capabilities
- `adaptive-idle-timeout`: Idle-based timeout mechanism that extends request lifetime on agent activity, with hard ceiling safety net and session history cleanup on timeout

### Modified Capabilities
- `auto-extend-timeout`: IdleTimeout config field supersedes AutoExtendTimeout; legacy AutoExtendTimeout continues to work via resolveTimeouts() mapping

## Impact

- `internal/deadline/` — new shared package
- `internal/app/deadline.go` — thin backward-compat wrapper
- `internal/app/channels.go` — refactored runAgent with resolveTimeouts()
- `internal/gateway/server.go` — idle timeout in handleChatMessage
- `internal/app/wiring.go` — pass IdleTimeout/MaxTimeout to gateway config
- `internal/config/types.go` — IdleTimeout field
- `internal/adk/errors.go` — ErrIdleTimeout error code
- `internal/session/store.go` — AnnotateTimeout interface method
- `internal/session/ent_store.go` — AnnotateTimeout implementation
- All session.Store mock implementations updated
