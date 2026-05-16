## Why

`zkexport --all` currently removes the file for the circuit that fails, but if a later circuit fails after earlier ones already succeeded, those earlier files are left behind. That makes the command only partially atomic and can leave a misleading mixed-state output directory.

## What Changes

- Track verifier files created during a `zkexport --all` run
- Remove those earlier files if a later circuit export fails
- Add regressions for mid-run cleanup
- Record the all-mode atomic cleanup contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: `zkexport --all` cleans up files created earlier in the same run when a later export fails

## Impact

- Affected code: `cmd/zkexport/main.go`, `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No surface change for successful runs; failure semantics become safer
