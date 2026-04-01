## Context

BudgetPolicy tracks turn and delegation counts in memory. When a session resumes (reconnect, process restart), these counters reset to zero. The existing `Session.Metadata` (map[string]string) provides a natural persistence point without schema changes.

MetricsCollector already tracks per-session token usage (input/output tokens) via SessionBreakdown, but this data is also in-memory only and lost on restart.

## Goals / Non-Goals

**Goals:**
- Persist budget counters (turns, delegations) into Session.Metadata after each turn
- Persist cumulative token usage (input/output tokens) into Session.Metadata after each turn
- Restore budget state lazily on first executor call per session after resume
- Zero-change to session schema, store interfaces, or ent models

**Non-Goals:**
- Persisting tool call counts (SessionMetric has no tool call count field)
- Real-time budget enforcement (BudgetPolicy is observational, not authoritative)
- Modifying the inner executor's hardcoded limits

## Decisions

### D1: Serialize/Restore on BudgetPolicy itself
Budget state serialization belongs on BudgetPolicy since it owns the mutable counters. Returns `map[string]string` matching Session.Metadata's type. Restore accepts the same map and silently ignores missing/malformed keys for forward compatibility.

**Alternative**: Separate serializer — rejected because it would need access to unexported fields.

### D2: Lazy restore via executor wrapper
A `budgetRestoringExecutor` wraps the real executor and restores budget state on first `RunStreamingDetailed` call per session. Uses `sync.Map` for the "already restored" check to avoid double-restore under concurrent calls.

**Alternative**: Eager restore at wiring time — rejected because the session key is not known at boot, only at runtime.

### D3: OnTurnComplete callback for persistence
After each turn, a callback serializes budget + token metrics into Session.Metadata and calls `store.Update()`. This piggybacks on the existing TurnRunner callback mechanism used by memory compaction and other buffers.

**Alternative**: EventBus subscriber — rejected because the existing pattern for post-turn side effects is OnTurnComplete callbacks.

### D4: Metadata key schema
Keys are prefixed with `usage:` to namespace within Session.Metadata:
- `usage:budget_turns` — strconv int
- `usage:budget_delegations` — strconv int
- `usage:cumulative_input_tokens` — strconv int
- `usage:cumulative_output_tokens` — strconv int

### D5: Return budget from initAgentRuntime
Change return type to `(turnrunner.Executor, *agentrt.BudgetPolicy)`. Classic mode returns `(innerExecutor, nil)`. The nil check in app.go gates all budget-related wiring.

## Risks / Trade-offs

- [Store.Update per turn] → Acceptable overhead: Session.Update already happens for history append; metadata is a small map addition.
- [Stale MetricsCollector data after restart] → The collector resets on restart, so only the persisted cumulative values survive. The callback merges both sources.
- [sync.Map memory for restored tracking] → Bounded by active session count, negligible.
