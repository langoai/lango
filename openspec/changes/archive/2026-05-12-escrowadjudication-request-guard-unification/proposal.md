## Why

The execution cluster now treats an empty transaction receipt id as a validation error in most entrypoints, but escrow adjudication still converts that input bug into a denied business result. That keeps one last operator-facing inconsistency in the adjudication path.

## What Changes

- Make escrow adjudication reject an empty transaction receipt id with an actionable validation error.
- Update the adjudication regression to match the unified execution-cluster contract.
- Sync `production-readiness` coverage for the adjudication request-id guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: escrow adjudication now treats an empty transaction receipt id as a validation error rather than a denied business outcome.

## Impact

- Affected code: `internal/escrowadjudication/service.go`
- Affected tests: `internal/escrowadjudication/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
