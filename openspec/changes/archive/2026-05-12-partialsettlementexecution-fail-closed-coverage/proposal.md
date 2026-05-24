## Why

The partial settlement execution service already rejects missing receipt-store and runtime dependencies, but those fail-closed guarantees are not directly protected by regression tests. That leaves a partial-settlement execution boundary weakly verified.

## What Changes

- Add regression tests for missing receipt-store and runtime wiring in `partialsettlementexecution.Service`.
- Sync `production-readiness` coverage for partial settlement execution dependency guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: partial settlement execution now has explicit regression coverage for missing-store and missing-runtime fail-closed behavior.

## Impact

- Affected tests: `internal/partialsettlementexecution/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
