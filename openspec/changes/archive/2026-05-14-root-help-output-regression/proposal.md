## Why

The bare `lango` root command already falls back to `cmd.Help()` in non-interactive mode, but that top-level help-routing contract is not locked down by a regression. A future change could accidentally bypass Cobra output routing without being noticed.

## What Changes

- Add a regression proving that non-interactive bare-root help writes to the Cobra command output stream
- Record the contract in the CLI reference spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: bare-root non-interactive help is command-stream routed

## Impact

- Affected code: `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
