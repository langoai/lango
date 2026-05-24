## Why

The `run_*` control-plane tools declare multiple required wrapper inputs, but several handlers still used optional extraction and could defer missing input into downstream snapshot or journal errors. That weakens the control-plane contract and makes failures less actionable.

## What Changes

- Switch `run_*` handlers to explicit required-input extraction for their declared required parameters.
- Add direct tool-entrypoint regressions for missing required inputs across the run-ledger tool surface.
- Update run-ledger feature docs and production/spec coverage for the fail-closed wrapper contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `run-ledger`: control-plane tools now fail closed on missing required wrapper inputs with direct regression coverage.
- `production-readiness`: wrapper guard coverage now includes the `run_*` control-plane tools.

## Impact

- Affected code: `internal/runledger/tools.go`
- Affected tests: `internal/runledger/tools_test.go`
- Affected docs: `docs/features/run-ledger.md`
- Affected specs: `openspec/specs/run-ledger/spec.md`, `openspec/specs/production-readiness/spec.md`
