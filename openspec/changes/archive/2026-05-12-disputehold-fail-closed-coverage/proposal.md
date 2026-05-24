## Why

The dispute-hold service already rejects missing receipt-store and runtime dependencies, but those fail-closed guarantees are not directly protected by regression tests. That leaves a dispute-critical execution boundary weakly verified.

## What Changes

- Add regression tests for missing receipt-store and runtime wiring in `disputehold.Service`.
- Sync `production-readiness` coverage for dispute-hold dependency guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: dispute hold now has explicit regression coverage for missing-store and missing-runtime fail-closed behavior.

## Impact

- Affected tests: `internal/disputehold/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
