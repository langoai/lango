## Context

The ADK translation layer in `internal/adk/` (~79KB, 7 files) bridges Lango's internal types with Google ADK v0.6.0. Over time, conversion logic for FunctionCall/FunctionResponse has been duplicated across two code paths:
- **Save direction**: `eventToMessage()` at `session_service.go:299-370`
- **Restore direction**: inline block within `EventsAdapter.All()` at `state.go:282-360+`

Four critical bug fixes are embedded inline without clear extraction. ADK v0.6.0 provides `plugin.Config`, `mcptoolset.New()`, and `memory.Service` that our `runner.Config` does not pass through (`agent.go:125-129`).

## Goals / Non-Goals

**Goals:**
- Consolidate FunctionCall/FunctionResponse conversion into shared converter functions with a single source of truth
- Build golden test suite as regression safety net before any structural changes
- Split `context_model.go` into focused files without changing logic
- Restructure `toolCallAccumulator` to provider-agnostic state machine
- Spike ADK plugin and MCPToolset integration feasibility with concrete adoption criteria
- Document `SessionServiceAdapter.Get()` contract deviation from ADK's `session.Service.Get()`

**Non-Goals:**
- Error classification type system (repo-wide scope, separate initiative)
- ADK type boundary isolation (depends on plugin spike results)
- Memory Service integration (parity gap too large for 1st batch)
- Tool double-adaptation removal (depends on MCPToolset spike results)
- Any CLI/TUI/config changes

## Decisions

### D1: Extract inline parts assembly before unifying converters
**Decision**: First extract `state.go:282-360+` inline block into a named function (`buildEventParts`), then identify shared logic with `eventToMessage()`.
**Rationale**: The restore direction has no function boundary — jumping straight to unification would require understanding both directions simultaneously. Extract-then-unify is safer and more reviewable.
**Alternative**: Rewrite both from scratch into a bidirectional converter. Rejected — too risky with 4 embedded bug fixes.

### D2: Keep Get() auto-create/renew behavior
**Decision**: Maintain `SessionServiceAdapter.Get()` auto-create and auto-renew, add contract-deviation comment and regression test.
**Rationale**: ADK runner calls `sessionService.Get()` directly at `runner.go:119`. Removing auto-create breaks the first turn. Adding explicit pre-create requires intercepting runner's internal flow, which is fragile.
**Alternative**: Add pre-create in `agent.Run()` before runner invocation. Rejected — runner.Run() calls Get() internally, creating race between our pre-create and runner's Get().

### D3: Provider-agnostic state machine for toolCallAccumulator
**Decision**: Replace OpenAI index-based / Anthropic ID-based branching with states: `Idle → Receiving(index/id) → Complete`. Orphaned delta = delta received in `Idle` state → drop with warning.
**Rationale**: Current code uses `hasAny` boolean and `lastIndex` tracking which obscures the state transitions. State machine makes the orphaned-delta invariant explicit.
**Alternative**: Keep current structure, just add comments. Rejected — the branching logic is the source of streaming bugs.

### D4: Plugin spike scope — agent-level vs per-tool granularity
**Decision**: Spike will verify but NOT adopt. The key gap is ADK callbacks are agent-level (all tools), while our toolchain middleware supports per-tool application. Spike documents this gap and identifies which middlewares CAN move (agent-level ones like logging) vs MUST stay (per-tool ones like approval).
**Rationale**: Premature adoption would lose per-tool granularity that approval, safety-level, and access-control depend on.

### D5: MCPToolset adoption criteria — all 5 conditions must pass
**Decision**: Adoption requires: (1) naming contract preservation, (2) approval path parity, (3) safety metadata propagation, (4) output truncation, (5) event publication. All must pass simultaneously.
**Rationale**: Partial adoption creates two tool execution paths with different security guarantees — unacceptable.

## Risks / Trade-offs

- **[Risk] Converter unification breaks subtle edge cases** → Mitigated by golden test suite (Unit 1) running before any structural changes. All 4 bug fix scenarios have explicit test cases.
- **[Risk] State machine refactor introduces new streaming regressions** → Mitigated by existing provider-specific streaming tests + new golden tests for orphaned delta and partial/final deduplication.
- **[Risk] Plugin spike creates pressure to adopt prematurely** → Mitigated by documenting spike as evaluation only, with explicit "do not adopt" for per-tool middlewares.
- **[Trade-off] Keeping legacy fallback path in state.go** → Increases code size but preserves backward compatibility for sessions created before Output metadata was added. Removal deferred to post-migration verification.
- **[Trade-off] context_model.go split is pure file organization** → No logic change, so minimal risk but also minimal functional improvement. Value is in readability and future modification safety.
