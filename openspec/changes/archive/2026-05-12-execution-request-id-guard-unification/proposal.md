## Why

The execution cluster still handles one basic operator mistake inconsistently. `settlementexecution` now returns a direct validation error for an empty transaction receipt id, but adjacent execution services still convert that input bug into a denied business result. That weakens operator-facing consistency across the same workflow family.

## What Changes

- Make escrow release, escrow refund, dispute hold, and partial settlement execution reject empty transaction receipt ids with actionable validation errors.
- Update regressions for those entrypoints to match the unified contract.
- Sync `production-readiness` coverage for the unified request-id guard behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: execution services in the settlement cluster now treat an empty transaction receipt id as a validation error rather than a denied business outcome.

## Impact

- Affected code: `internal/escrowrelease/service.go`, `internal/escrowrefund/service.go`, `internal/disputehold/service.go`, `internal/partialsettlementexecution/service.go`
- Affected tests: `internal/escrowrelease/service_test.go`, `internal/escrowrefund/service_test.go`, `internal/disputehold/service_test.go`, `internal/partialsettlementexecution/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
