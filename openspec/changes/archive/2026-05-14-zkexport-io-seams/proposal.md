## Why

`cmd/zkexport` currently parses global flags directly in `main()` and writes usage, errors, and progress through raw process stdio. That makes the utility hard to exercise in tests and forces automation to treat it as a black box.

## What Changes

- Introduce a testable `runZKExport(args, stdout, stderr)` path with local flag parsing
- Add a prover/export seam so success and failure branches can be tested without running real Groth16 setup
- Add regressions for usage, single-circuit success, `--all` success, and prover-service failure
- Record the output-routing contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: zkexport success output writes to stdout and usage/errors write to stderr

## Impact

- Affected code: `cmd/zkexport/main.go`, `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No behavior change for operators; this is utility testability and IO-contract hardening
