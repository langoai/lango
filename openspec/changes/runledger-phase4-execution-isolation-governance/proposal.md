## Why

Phases 1~3 establish correctness, durability, and authoritative state. The final missing piece
is narrowing execution surface area so the runtime can safely run coding tasks with real
workspace isolation and step-scoped tool exposure.

## What Changes

- Activate workspace isolation in the app runtime behind the phase gate.
- Enforce step-scoped tool profiles during execution.
- Tighten execution ownership and policy-application paths around isolated runs.
- Update docs and operator guidance for fail-closed isolated execution.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `run-ledger`: production workspace isolation activation and step-scoped tool governance

## Impact

- `internal/app/modules_runledger.go`
- `internal/runledger/pev.go`
- `internal/runledger/workspace.go`
- `internal/orchestration/`
- `internal/toolcatalog/`
- `openspec/specs/run-ledger/spec.md`
