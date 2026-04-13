## Context

After merging 11 feature branches into dev (1,073 files, +110K lines), code analysis identified 12 refactoring targets across 4 priority levels. All changes are internal — no API or config surface changes. Implementation is complete; this design documents the decisions made.

## Goals / Non-Goals

**Goals:**
- Eliminate TUI chat hot-path performance bottleneck (glamour renderer recreation per tick)
- Achieve O(1) RunSnapshot.FindStep() via lazy step index
- Enable future reconciliation by persisting workflow/background descriptors in journal
- Reduce store boilerplate with shared helpers
- Standardize tool parameter extraction across all tools.go files
- Extract event name constants to prevent stringly-typed bugs

**Non-Goals:**
- Callback→EventBus migration (domain-internal hooks stay per feedback_callback_scope.md)
- Generic base Store abstraction (storeutil helpers are sufficient)
- Reconciler goroutine (descriptor persistence is prerequisite; reconciler is separate change)
- Finding.SearchSource rename (blocked by agentic-retrieval spec contract)
- TurnTrace nil→error change (blocked by agent-turn-tracing spec contract)
- FTS5 startup reindex skip (blocked by knowledge-fts5-integration spec contract)

## Decisions

1. **Glamour renderer: module-level cache vs per-entry cache** → Module-level width-keyed cache. Bubbletea is single-threaded so no synchronization needed. Simpler than caching per transcriptItem.

2. **FindStep index: maintained field vs lazy init** → Lazy init (`ensureStepIndex()`). RunSnapshot is a plain struct used in JSON unmarshal/DeepCopy/applyEvent. Maintained field would require 5+ synchronization points. Lazy init has 2 invalidation points (PlanAttached, PolicyDecompose) and safely rebuilds on next access.

3. **storeutil.MarshalField: error swallowing vs fail-fast** → Fail-fast `(json.RawMessage, error)`. Initial implementation swallowed errors (matching `marshalPayload` pattern), but store persistence requires fail-fast to prevent silent data corruption. `marshalPayload` is only for best-effort log payloads.

4. **SourceKind discriminator: type field vs JSON shape inference** → Explicit `SourceKind string` field in `RunCreatedPayload`. Avoids brittle JSON shape heuristics in future reconciler. Values: "workflow", "background", "" (legacy).

5. **Coordinator merge: mutex vs pre-allocated index** → Pre-allocated `[][]Finding` indexed by agent position. Each goroutine writes to its own index — no lock needed. Flatten after `g.Wait()`.

6. **Contextpanel sort guard: dirty flag vs content hash** → Dirty flag with lightweight change detection (tool count sum + map length comparison). Full content hash is overkill for a 5-second refresh interval.

## Risks / Trade-offs

- [Lazy step index stale data] → Mitigated by invalidation at exactly 2 mutation points + 4 test scenarios covering PlanAttached/PolicyDecompose/DeepCopy/JSON rehydrate
- [storeutil adoption is partial] → Only 2 store files migrated; others use storeutil in future changes. No regression since existing code untouched.
- [toolparam migration incomplete] → 7/10 files migrated. Remaining 3 (tools_meta, tools_smartaccount, tools_contract) have complex conditional patterns; deferred to avoid scope creep.
- [SourceDescriptor backward compat] → `json:",omitempty"` ensures legacy journals without the field unmarshal cleanly to zero value.
