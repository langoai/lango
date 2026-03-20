## Why

After persistence and write-through land, the next gap is read-path consistency. As long as
workflow/background/gateway read from mixed sources, RunLedger is not the true operational
authority. Phase 3 makes snapshots authoritative for reads and turns resume from a local helper
into a system-level behavior.

## What Changes

- Promote RunLedger snapshots to the authoritative read path.
- Integrate opt-in resume candidate discovery with gateway/agent context.
- Add Command Context injection based on active run summaries.
- Harden pause/resume semantics around interrupted and resumable runs.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `run-ledger`: authoritative reads, command context injection, and explicit resume orchestration

## Impact

- `internal/runledger/resume.go`
- `internal/runledger/snapshot.go`
- `internal/adk/context_model.go`
- `internal/gateway/`
- `internal/app/wiring.go`
- `openspec/specs/run-ledger/spec.md`
