## Why

Lango's P2P layer lacks session-level provenance tracking — there is no way to record checkpoints during task execution, track session hierarchies (fork/merge/discard), or attribute contributions across agents and humans. Adopting session/checkpoint/attribution concepts (inspired by Entire.io's model) enables auditable execution history, reproducible states via journal replay, and fair contribution tracking for the economy layer.

## What Changes

- **Checkpoint system**: Thin metadata records (run_id + journal_seq + label + trigger) that mark points in RunLedger journals. Manual and automatic creation via RunLedger append hooks.
- **RunLedger append hook**: `StoreOption`-based callback system for both `MemoryStore` and `EntStore`, enabling decoupled consumers to react to journal events without modifying the `RunLedgerStore` interface.
- **Session tree**: Hierarchical tracking of session lifecycle events (fork/merge/discard) via `InMemoryChildStore` lifecycle hooks.
- **Attribution framework**: Coarse-grained contribution tracking (member/file/commit/session level) with report generation (Phase 3).
- **Provenance bundle**: Portable container format with 3-level redaction for P2P transport (Phase 4).
- **CLI**: `lango provenance status|checkpoint|session|attribution` command group.
- **Config**: `provenance.enabled`, `provenance.checkpoints.*` configuration section.

## Capabilities

### New Capabilities
- `session-provenance`: Core provenance system — checkpoints, session tree, attribution, bundles, and CLI

### Modified Capabilities
- `run-ledger`: Add `StoreOption` / `WithAppendHook` to `MemoryStore` and `EntStore` constructors (backward-compatible variadic)
- `config-system`: Add `ProvenanceConfig` section to root config with checkpoint settings
- `cli-status-dashboard`: Add Provenance feature line to status output

## Impact

- **Core packages**: `internal/provenance/` (new), `internal/runledger/` (options.go + store modifications), `internal/session/` (child_store lifecycle hook), `internal/config/` (types_provenance.go)
- **CLI**: `internal/cli/provenance/` (new command group), `internal/cli/status/` (feature line), `cmd/lango/main.go` (registration)
- **App wiring**: `internal/app/modules_provenance.go` (new module), `internal/app/app.go` (module registration)
- **Ent schemas**: `provenance_checkpoint`, `session_provenance` (new tables)
- **No breaking changes**: All store constructor changes use variadic options for backward compatibility
