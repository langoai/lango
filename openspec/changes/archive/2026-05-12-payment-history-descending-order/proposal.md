## Why

The production-readiness spec already requires payment history to return records in descending order with a limit, but the existing tests only verified count and field mapping. That left the ordering part of the contract weakly enforced.

## What Changes

- Add a payment history regression that seeds a newer transaction and verifies it sorts first.
- Assert that `History(limit=1)` trims on top of the descending ordering contract rather than on insertion order.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: Payment history ordering is now enforced by executable tests rather than implied by query construction.

## Impact

- Affected tests: `internal/payment/service_test.go`
