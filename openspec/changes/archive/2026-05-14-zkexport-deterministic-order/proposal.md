## Why

`zkexport` currently derives its circuit list from a Go map. Without explicit sorting, the usage text and `--all` export progress order can vary, which makes automation output noisier and regressions harder to reason about.

## What Changes

- Sort `zkexport` circuit IDs before rendering usage text
- Sort `zkexport --all` export iteration so progress output is deterministic
- Add regressions for usage ordering and all-mode progress ordering
- Record the deterministic ordering contract in the ZKP core spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `zkp-core`: zkexport circuit listing and all-mode export order are deterministic

## Impact

- Affected code: `cmd/zkexport/main.go`, `cmd/zkexport/main_test.go`
- Affected specs: `openspec/specs/zkp-core/spec.md`
- No behavior change beyond stabilizing output order
