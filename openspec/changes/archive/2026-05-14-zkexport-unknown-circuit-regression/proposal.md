## Why

`zkexport` already reports unknown circuit IDs with the available circuit list, but that actionable failure path is not explicitly covered by regression tests. Since the available list is now deterministic, it is worth locking the full error text down.

## What Changes

- Add a regression for unknown-circuit failure
- Assert that the available-circuit list in the stderr message is deterministic
- Record the contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: unknown circuit failures include a deterministic available-list stderr message

## Impact

- Affected code: `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No runtime behavior changes
