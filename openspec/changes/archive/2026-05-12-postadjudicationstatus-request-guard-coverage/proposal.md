## Why

The post-adjudication status service currently treats an empty transaction receipt id like a missing record. That blurs a basic operator input mistake into a not-found path and weakens the status-read contract.

## What Changes

- Make `GetTransactionStatus` reject an empty transaction receipt id with an actionable validation error.
- Add regressions for that request guard.
- Sync `production-readiness` coverage for the status request-id validation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: post-adjudication status now has explicit validation coverage for missing transaction receipt ids.

## Impact

- Affected code: `internal/postadjudicationstatus/types.go`, `internal/postadjudicationstatus/service.go`
- Affected tests: `internal/postadjudicationstatus/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
