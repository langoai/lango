## Why

`lango` exposes a root-level persistent `--mode` flag for interactive entrypoints. Utility subcommands such as `version` and `health` should ignore that flag cleanly, but there is no regression that proves the shared parser behavior stays harmless for those commands.

## What Changes

- Add a regression for `lango version --mode <value>`
- Add a regression for `lango health --mode <value>`
- Record the contract in the CLI reference spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: root `--mode` does not interfere with top-level utility subcommands

## Impact

- Affected code: `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
