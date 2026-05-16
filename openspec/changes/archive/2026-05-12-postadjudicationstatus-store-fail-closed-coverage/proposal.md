## Why

The post-adjudication status service already has explicit store guards, but those fail-closed contracts are not directly enforced by regression tests. That leaves an operator-facing status boundary weakly verified.

## What Changes

- Add regressions for missing receipt-store wiring across post-adjudication status entrypoints.
- Sync `production-readiness` coverage for the status service dependency guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: post-adjudication status entrypoints now have explicit regression coverage for missing-store fail-closed behavior.

## Impact

- Affected tests: `internal/postadjudicationstatus/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
