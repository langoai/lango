## Why

The post-adjudication replay service already rejects empty `transaction_receipt_id` input, but that request-validation contract is not directly protected by regression tests. That leaves a user-facing replay guard weakly verified.

## What Changes

- Add a regression test for missing `transaction_receipt_id` in `postadjudicationreplay.Service`.
- Sync `production-readiness` coverage for replay request validation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: post-adjudication replay now has explicit regression coverage for missing transaction receipt id validation.

## Impact

- Affected tests: `internal/postadjudicationreplay/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
