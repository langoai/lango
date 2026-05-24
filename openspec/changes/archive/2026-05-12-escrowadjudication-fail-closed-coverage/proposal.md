## Why

The escrow adjudication service already rejects a missing receipt store, but that fail-closed guarantee is not directly protected by regression tests. That leaves a dispute-critical adjudication boundary weakly verified.

## What Changes

- Add a regression test for missing receipt-store wiring in `escrowadjudication.Service`.
- Sync `production-readiness` coverage for the adjudication dependency guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: escrow adjudication now has explicit regression coverage for missing-store fail-closed behavior.

## Impact

- Affected tests: `internal/escrowadjudication/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
