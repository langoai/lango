## Context

The adaptive-idle-timeout feature was implemented in a prior change. Code review identified that `initGateway()` in `wiring.go` had a simplified inline copy of timeout resolution logic that missed several cases handled by `channels.go`'s `resolveTimeouts()`. Additionally, `gateway/server.go` used raw string literals for timeout error types rather than the typed `deadline.Reason` constants.

## Goals / Non-Goals

**Goals:**
- Single source of truth for timeout resolution logic via `deadline.ResolveTimeouts()`
- Eliminate logic divergence between channel and gateway timeout handling
- Replace raw string error types with typed constants in gateway

**Non-Goals:**
- Changing timeout behavior or defaults
- Refactoring `ExtendableDeadline` internals (timer lifecycle, mutex strategy)
- Adding new timeout features

## Decisions

### Extract `ResolveTimeouts` to `deadline` package

**Decision**: Create `deadline.TimeoutConfig` struct and `deadline.ResolveTimeouts(cfg) (idle, ceiling)` as a pure function. Both `channels.go` and `wiring.go` call this function.

**Rationale**: The `deadline` package already owns `ExtendableDeadline` and `Reason` constants. Timeout resolution is logically part of the same domain. A pure function with a config struct is easily testable without `App` dependencies.

**Alternative considered**: Keep `resolveTimeouts()` as an `App` method and call it from `initGateway()`. Rejected because `initGateway()` is a package-level function, not a method, and passing the full `App` would increase coupling.

### Use `string(deadline.ReasonXxx)` in gateway error types

**Decision**: Replace `"timeout"`, `"idle_timeout"`, `"max_timeout"` raw strings with `string(deadline.ReasonIdle)` and `string(deadline.ReasonMaxTimeout)`.

**Rationale**: The `deadline.Reason` type already defines these exact constants. Using them prevents typo-based divergence and makes the gateway error types traceable to their source.

## Risks / Trade-offs

- **[Risk] Gateway default errType changes from `"timeout"` to `"max_timeout"`** → This only affects the WebSocket `agent.error` event payload. The `"timeout"` string was generic; `"max_timeout"` is more precise. UI clients should already handle unknown types gracefully.
- **[Risk] Test relocation may confuse blame history** → Mitigated by keeping the integration test in `channels_test.go` to confirm delegation works.
