## Why

The escrow release service already rejects missing receipt-store and runtime dependencies, but those fail-closed guarantees are not directly protected by regression tests. That leaves a settlement-critical execution boundary weakly verified.

## What Changes

- Add regression tests for missing receipt-store and escrow-runtime wiring in `escrowrelease.Service`.
- Sync `production-readiness` coverage for escrow release dependency guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: escrow release now has explicit regression coverage for missing-store and missing-runtime fail-closed behavior.

## Impact

- Affected tests: `internal/escrowrelease/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
