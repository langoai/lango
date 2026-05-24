## Why

The mission service already rejects missing `source_kind` and `execution_ref` inputs, but those operator-facing validation contracts are not directly protected by regression tests. That leaves core mission workflow validation weakly verified.

## What Changes

- Add regression tests for missing `source_kind` on accepted proposals and missing `execution_ref` on execution links.
- Sync `production-readiness` coverage for mission request validation.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: mission service request validation now has explicit regression coverage for required proposal/execution fields.

## Impact

- Affected tests: `internal/mission/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
