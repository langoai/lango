## Context

Lango's RunLedger provides append-only journal-based durable execution, but lacks cross-cutting provenance tracking. Sessions fork/merge/discard without recording their lineage. There is no checkpoint mechanism to mark significant journal positions. Attribution of work across agents and humans is not tracked.

The provenance system layers on top of existing infrastructure (RunLedger journals, session child stores) using callback hooks — no core interface changes required.

## Goals / Non-Goals

**Goals:**
- Checkpoint creation (manual + automatic) as thin metadata referencing journal positions
- Session tree tracking via lifecycle hooks on InMemoryChildStore
- RunLedger append hook mechanism for decoupled event consumption
- CLI for provenance inspection (`lango provenance`)
- Config-driven auto-checkpoint behavior
- Foundation for Phase 3 (attribution) and Phase 4 (P2P bundle transport)

**Non-Goals:**
- Full snapshot storage in checkpoints (use journal replay instead)
- Rewind/mutation of existing runs (v1 uses fork-from-checkpoint concept only)
- Line-level attribution (coarse member/file/commit granularity only)
- ZK proof of provenance (Phase 5)
- Entire.io CLI/format dependency

## Decisions

### 1. Thin Checkpoint Metadata (not full snapshots)
Checkpoints store `run_id + journal_seq + label + trigger` — no snapshot data. Restoration replays the journal up to `journal_seq` using the existing `MaterializeFromJournal` path.

**Rationale**: RunSnapshot is a projection derived from journal replay. Storing full snapshots in checkpoints would duplicate the source of truth and violate the append-only journal philosophy.

**Alternative**: Store serialized RunSnapshot — rejected because it creates divergence risk and storage bloat.

### 2. StoreOption + AppendHook (not EventBus bridge)
RunLedger stores accept `WithAppendHook(func(JournalEvent))` as a functional option. The hook is called synchronously after successful append, outside the store lock.

**Rationale**: No RunLedgerStore interface change (interfaces are sacred per teammate rules). Functional options pattern matches existing codebase conventions. Callback runs outside lock to prevent deadlocks when the hook reads back from the store.

**Alternative**: EventBus bridge — rejected because RunLedger journal events are not on the EventBus, and adding them creates a new coupling point.

### 3. ChildStoreOption + WithLifecycleHook
Same functional option pattern applied to `InMemoryChildStore` for fork/merge/discard lifecycle tracking.

**Rationale**: Consistent with the RunLedger hook pattern. Decoupled from provenance package (hook is a plain `func` type, no import cycle).

### 4. In-Memory Stores First (Ent later)
Phase 1 uses `MemoryStore` / `MemoryTreeStore` for checkpoint and session tree persistence. Ent-backed stores use the new schemas but are wired in Phase 2.

**Rationale**: Reduces initial complexity. Ent schemas are created now (for DB migration), but wiring goes through in-memory stores for testing and iteration speed.

### 5. Module System Integration
Provenance is an `appinit.Module` with `DependsOn: [ProvidesRunLedger]`. It resolves the RunLedger store from the module resolver to wire the checkpoint service.

**Rationale**: Follows established module pattern. Dependency on RunLedger is explicit and topologically sorted.

## Risks / Trade-offs

- **[Hook ordering]** AppendHook is synchronous — slow hooks block journal appends. → Mitigation: Document that hooks must be lightweight. Future: async hook option.
- **[Memory store data loss]** In-memory stores lose data on restart. → Mitigation: Phase 2 adds Ent-backed stores. Phase 1 is for establishing the data model.
- **[Schema migration]** New Ent schemas add tables on bootstrap. → Mitigation: Ent auto-migration is additive-only; no data loss risk.
- **[Backward compatibility]** Variadic `StoreOption` changes are source-compatible but not binary-compatible. → Mitigation: All callers within the same module; no external consumers.
