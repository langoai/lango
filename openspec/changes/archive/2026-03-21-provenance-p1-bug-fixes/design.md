## Context

Session Provenance was implemented in the previous change (`session-provenance`). Code review found three bugs that prevent real-world usage:

1. **Seq=0 in hook**: `EntStore.AppendJournalEvent` declares `nextSeq` inside a closure, so the hook always receives `event.Seq=0`. MemoryStore correctly assigns `event.Seq = seq`.
2. **No auto-checkpoint wiring**: `WithAppendHook` only works at construction time. The app module system creates RunLedger store first, then Provenance module — so there's no way to register the hook post-construction.
3. **Ephemeral CLI stores**: CLI commands create `NewMemoryStore()` on every invocation, losing all checkpoint data between process runs.

## Goals / Non-Goals

**Goals:**
- Fix Seq propagation so hooks receive correct monotonic sequence numbers
- Enable post-construction hook registration with chaining (no overwrite)
- Persist checkpoints via Ent-backed store (same DB as RunLedger)
- Wire auto-checkpoint hook in app module initialization

**Non-Goals:**
- Persistent session tree store (separate follow-up)
- CLI E2E tests (package-level integration tests sufficient)
- Changes to the `RunLedgerStore` interface

## Decisions

### 1. Hoist `nextSeq` out of closure
**Decision**: Move `var nextSeq int64` before the retry loop; change `nextSeq := int64(1)` to `nextSeq = int64(1)` (assignment, not declaration). Add `event.Seq = nextSeq` before hook call.
**Rationale**: Minimal change, matches MemoryStore behavior exactly. The variable must survive the closure scope to be visible at hook call site.

### 2. `AppendHookSetter` narrow interface (not modifying `RunLedgerStore`)
**Decision**: Add `AppendHookSetter` interface in `options.go` with single method `SetAppendHook(func(JournalEvent))`. Implement on concrete types only.
**Rationale**: The `RunLedgerStore` interface is a published contract. Adding a setter there would force all implementations to support it. A narrow interface + type assertion (`if setter, ok := store.(AppendHookSetter)`) keeps the contract stable.
**Alternative considered**: Adding `SetAppendHook` to `RunLedgerStore` — rejected because it's a wiring concern, not a store concern.

### 3. Hook chaining (not replacement)
**Decision**: `SetAppendHook` chains with existing hooks: `func(e) { prev(e); h(e) }`.
**Rationale**: Multiple modules may register hooks. Replacement would silently break earlier registrations. Chaining is safe because hooks are registered during sequential boot (no concurrency concern).

### 4. `EntCheckpointStore` reusing existing Ent schema
**Decision**: Implement `CheckpointStore` interface backed by `ent.ProvenanceCheckpoint` table (schema already exists from session-provenance change).
**Rationale**: Schema and migration already in place. Just needs the Go store implementation.

### 5. Session CLI → "not yet implemented" placeholder
**Decision**: Replace tree/list commands with honest placeholder messages instead of silently returning empty results from ephemeral stores.
**Rationale**: Empty results from a fresh MemoryStore are misleading — users think they have no sessions when in fact the store is just ephemeral. An explicit message sets correct expectations.

## Risks / Trade-offs

- **[Risk]** `SetAppendHook` called after concurrent `AppendJournalEvent` starts → **Mitigation**: Document "must be called during boot, before concurrent access". Boot is sequential by design.
- **[Risk]** EntCheckpointStore metadata JSON round-trip loses empty maps → **Mitigation**: Only marshal when `len(Metadata) > 0`; nil map on read is acceptable (matches MemoryStore contract).
