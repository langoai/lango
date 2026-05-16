## Why

`cmd/zkexport` already routes behavior through `runZKExport(...)`, but the top-level `main()` function still binds directly to `os.Args`, `os.Stdout`, `os.Stderr`, and `os.Exit`. That makes the final wrapper impossible to regression-test directly.

## What Changes

- Add injected args/stdout/stderr/exit seams for the `zkexport` main wrapper
- Add a regression showing that `main()` forwards the usage failure path into the injected stderr and exit seam
- Record the wrapper-seam contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: zkexport main wrapper is seam-aware and testable

## Impact

- Affected code: `cmd/zkexport/main.go`, `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No behavior change for operators; this is wrapper-level testability hardening
