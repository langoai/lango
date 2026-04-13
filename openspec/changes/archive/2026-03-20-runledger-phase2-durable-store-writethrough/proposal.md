## Why

RunLedger is still in the Phase 1 scaffold/hardening stage centered on `MemoryStore`. For Task OS to become the actual source of truth, journal/snapshot/step projection must survive beyond process lifecycle, and workflow/background write paths must go through the ledger first. Without this stage, RunLedger cannot transition from "library" to "runtime authority".

## What Changes

- Introduce an Ent-backed persistent store to RunLedger.
- RunLedger becomes the single source for `run_id` generation, and write-through adapters are added so that workflow/background projection uses the same ID.
- Add projection drift detection and replay/rebuild paths.
- Extend CLI so `lango run journal` can read from the persistent store.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `run-ledger`: persistent store, canonical run ID ownership, write-through projections,
  and projection rebuild/drift handling

## Impact

- `internal/runledger/store.go`
- `internal/runledger/writethrough.go`
- `internal/runledger/journal.go`
- `internal/runledger/snapshot.go`
- `internal/background/manager.go`
- `internal/workflow/state.go`
- `internal/cli/run/run.go`
- `openspec/specs/run-ledger/spec.md`
