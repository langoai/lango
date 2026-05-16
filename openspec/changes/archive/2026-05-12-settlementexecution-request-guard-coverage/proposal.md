## Why

The settlement execution service already rejects empty transaction receipt identifiers, but that user-facing validation contract is not directly protected by regression tests. That leaves a core execution entrypoint weakly verified for basic operator mistakes.

## What Changes

- Add a regression test for missing `transaction receipt id` in `settlementexecution.Service`.
- Sync `production-readiness` coverage for the settlement execution request guard.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `production-readiness`: settlement execution now has explicit regression coverage for missing transaction receipt id validation.

## Impact

- Affected tests: `internal/settlementexecution/service_test.go`
- Affected specs: `openspec/specs/production-readiness/spec.md`
