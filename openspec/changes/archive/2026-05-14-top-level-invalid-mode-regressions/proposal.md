## Why

The top-level interactive entrypoints already validate `--mode` against the configured mode registry, but there is no dedicated regression proving that bad values fail early with an actionable error across the root workbench, cockpit, and chat entrypoints.

## What Changes

- Add invalid-mode regressions for bare `lango`
- Add invalid-mode regressions for `lango cockpit`
- Add invalid-mode regressions for `lango chat`
- Record the contract in the CLI reference spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: top-level interactive entrypoints reject unknown `--mode` values consistently

## Impact

- Affected code: `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
