## Context

The session TTL tests (`TestEntStore_TTL`, `TestEntStore_TTL_DeleteAndRecreate`) use a 1ms TTL with a 5ms sleep. On Ubuntu CI with `-race` detector overhead, the window between `Create()` and `Get()` after session recreation exceeds 1ms, causing spurious `ErrSessionExpired` failures.

## Goals / Non-Goals

**Goals:**
- Eliminate flaky TTL test failures on Ubuntu CI with `-race` flag
- Maintain meaningful TTL expiration testing

**Non-Goals:**
- Changing TTL logic or production code
- Adding new test cases

## Decisions

**Decision: Use 50ms TTL with 100ms sleep**
- 50ms gives ample headroom for `Create` → `Get` (typically < 1ms even on slow CI)
- 100ms sleep provides 2x margin over TTL for reliable expiration
- Alternative considered: `time.AfterFunc` or polling — rejected as over-engineering for a simple timing fix
- Alternative considered: `t.Skip` on CI — rejected as it hides real test coverage

## Risks / Trade-offs

- [Slightly slower tests] → Adds ~190ms total across both tests; negligible impact on test suite duration
