## Why

The settlement progression service already returns actionable request/store validation errors, but those guards are not directly protected by regression tests. That leaves the operator-facing validation contract weakly verified.

## What Changes

- Add regressions for missing `transaction_receipt_id` and missing receipt-store wiring in `settlementprogression.Service`.
- Sync `production-readiness` coverage for those validation guards.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: settlement progression request/store validation guards now have explicit regression coverage.

## Impact

- Affected tests: `internal/settlementprogression/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
