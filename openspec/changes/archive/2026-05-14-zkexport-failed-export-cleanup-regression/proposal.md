## Why

`exportCircuit(...)` already removes partially-written output files on exporter failure, but that cleanup contract is not currently locked down by regressions. If a future refactor drops the cleanup, failed exports could leave misleading partial verifier files behind.

## What Changes

- Add a regression for single-circuit exporter failure cleanup
- Add a regression for `--all` exporter failure cleanup
- Record the cleanup and stderr-reporting contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: failed exports remove partial output and report the failure on stderr

## Impact

- Affected code: `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No runtime behavior changes
