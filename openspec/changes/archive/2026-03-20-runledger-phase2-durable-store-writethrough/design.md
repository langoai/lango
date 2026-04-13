## Context

Phase 1 made RunLedger testable and locally correct, but it still depends on `MemoryStore`
and does not own the write path for workflow/background execution. Phase 2 turns the ledger
into the actual durable authority layer.

## Goals

- Replace `MemoryStore` with an Ent-backed implementation for real persistence.
- Make `run_id` canonical and shared across all projections.
- Introduce write-through adapters so workflow/background writes hit RunLedger first.
- Make projection state rebuildable and drift-detectable.

## Non-Goals

- Switching read paths to authoritative snapshots everywhere.
- Enabling workspace isolation in production runtime.
- Implementing user-facing resume orchestration.

## Decisions

### 1. RunLedger owns `run_id`

Projection stores must not create their own IDs. The ledger creates the canonical run ID,
and workflow/background projections store mirrors keyed by that same ID.

### 2. Write-through before projection sync

The safe ordering is:

1. append journal event(s) to RunLedger
2. materialize/update snapshot
3. sync projection
4. append `projection_synced` marker

If projection sync fails, the run remains valid in RunLedger and can be repaired by replay.

### 3. Drift handling is explicit

Phase 2 adds:

- drift detection API
- rebuild-from-ledger projection repair
- degraded projection status instead of silent divergence

## Risks / Trade-offs

- Dual writes increase complexity, but replayable projections make failures recoverable.
- Ent-backed journal append must preserve per-run sequence monotonicity under concurrency.
- CLI journal viewing becomes useful only after persistence lands, so phase ordering matters.
