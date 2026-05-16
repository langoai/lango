## Why

The single-circuit `zkexport` path currently tries to build the prover service before checking whether the requested circuit ID even exists. That adds unnecessary setup work and can produce the wrong top-level failure if service initialization itself is broken.

## What Changes

- Validate the single-circuit ID before prover service setup
- Add a regression proving unknown circuit handling does not call the prover-service seam
- Extend the ZKP core spec with the short-circuit contract

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: unknown circuit failures short-circuit before prover service setup

## Impact

- Affected code: `cmd/zkexport/main.go`, `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No surface change for valid circuits; invalid-circuit failures become cleaner and cheaper
