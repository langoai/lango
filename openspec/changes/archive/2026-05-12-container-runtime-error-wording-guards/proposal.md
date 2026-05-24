## Why

The sandbox runtime tests already prove that unavailable container paths return `ErrRuntimeUnavailable`, but they do not directly guard the human-facing wording that explains which fail-closed path triggered the error. That leaves room for regressions where the operator still gets the right error type but loses the actionable reason.

## What Changes

- Add regression coverage for explicit Docker unavailable wording.
- Add regression coverage for explicit gVisor unavailable wording.
- Add regression coverage for the fail-closed `requireContainer` unavailable wording.
- Sync the production-readiness spec to treat those messages as part of the contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: sandbox runtime unavailability now has direct actionable-wording regressions for explicit Docker, explicit gVisor, and require-container fail-closed paths.

## Impact

- Affected code: `internal/sandbox/container_executor_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
