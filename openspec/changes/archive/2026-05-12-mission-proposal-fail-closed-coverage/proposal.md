## Why

The mission and proposal services already have explicit fail-closed dependency guards, but those contracts are not directly locked by regression tests. That leaves a coverage gap around core service wiring behavior even though the architecture intends to fail closed.

## What Changes

- Add regression tests for missing mission store dependencies across mission lifecycle entrypoints.
- Add regression tests for missing proposal registry and preparer dependencies across proposal service entrypoints.
- Sync `production-readiness` coverage for those fail-closed contracts.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: mission and proposal service dependency guards are now enforced by explicit regressions.

## Impact

- Affected tests: `internal/mission/service_test.go`, `internal/proposal/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
