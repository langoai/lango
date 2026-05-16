## Why

`zkexport` already special-cases `flag.ErrHelp`, but there is no regression proving that the help path returns success and avoids expensive prover setup. That leaves a small but important utility-command contract unguarded.

## What Changes

- Add a regression for `zkexport --help`
- Verify that the help path returns exit code `0`
- Verify that the prover-service seam is not invoked
- Record the contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: `zkexport --help` is regression-covered as a success path that skips prover setup

## Impact

- Affected code: `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No runtime behavior changes
