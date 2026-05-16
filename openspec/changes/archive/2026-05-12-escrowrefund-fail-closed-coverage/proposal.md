## Why

The escrow refund service already rejects missing receipt-store and runtime dependencies, but those fail-closed guarantees are not directly protected by regression tests. That leaves a settlement-critical refund boundary weakly verified.

## What Changes

- Add regression tests for missing receipt-store and refund-runtime wiring in `escrowrefund.Service`.
- Sync `production-readiness` coverage for escrow refund dependency guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: escrow refund now has explicit regression coverage for missing-store and missing-runtime fail-closed behavior.

## Impact

- Affected tests: `internal/escrowrefund/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
