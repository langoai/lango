## Why

Session Provenance (P1) has three bugs discovered in code review that prevent real-world usage despite passing tests: checkpoint CLI creates ephemeral stores losing all data on process exit, auto-checkpoint hooks are never wired at runtime, and EntStore passes Seq=0 to hooks due to a closure variable shadowing bug.

## What Changes

- Fix `EntStore.AppendJournalEvent` to hoist `nextSeq` out of the closure and assign `event.Seq = nextSeq` before calling the hook
- Add `AppendHookSetter` interface and `SetAppendHook` method on both `MemoryStore` and `EntStore` with hook chaining
- Wire `CheckpointService.OnJournalEvent` via `SetAppendHook` in `modules_provenance.go`
- Implement `EntCheckpointStore` (Ent-backed `CheckpointStore`) for persistent checkpoints
- Switch CLI checkpoint commands and app module from `NewMemoryStore()` to `NewEntCheckpointStore()`
- Replace session tree/list CLI commands with "not yet implemented" placeholder (pending persistent session tree store)

## Capabilities

### New Capabilities

_(none — all fixes are to existing capabilities)_

### Modified Capabilities

- `session-provenance`: Fix checkpoint persistence, auto-checkpoint wiring, and journal seq propagation
- `run-ledger`: Add post-construction hook registration via `AppendHookSetter` interface

## Impact

- `internal/runledger/ent_store.go` — Seq fix + `SetAppendHook`
- `internal/runledger/store.go` — `SetAppendHook` on MemoryStore
- `internal/runledger/options.go` — `AppendHookSetter` interface
- `internal/provenance/ent_store.go` — New `EntCheckpointStore`
- `internal/app/modules_provenance.go` — EntCheckpointStore + hook wiring
- `internal/cli/provenance/checkpoint.go` — Use EntCheckpointStore
- `internal/cli/provenance/session.go` — Placeholder for tree/list
