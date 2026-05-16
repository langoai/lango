## Why

The escrow execution service already rejects an empty transaction receipt id, but that validation contract is not directly protected by regression tests. That leaves a core escrow execution entrypoint weakly verified for a basic operator mistake.

## What Changes

- Add a regression test for missing `transaction receipt id` in `escrowexecution.Service`.
- Sync `production-readiness` coverage for the escrow execution request-id guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: escrow execution now has explicit regression coverage for missing transaction receipt id validation.

## Impact

- Affected tests: `internal/escrowexecution/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
