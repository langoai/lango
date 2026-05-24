## Why

The `runMain()` entrypoint gives sandbox worker mode highest priority, but that branch is not currently covered by a focused regression. A future refactor could accidentally let broker mode or root command setup run first.

## What Changes

- Add a `cmd/lango` regression for sandbox worker short-circuit behavior
- Assert that broker mode and root command construction are skipped when worker mode is active
- Record the contract in CLI reference spec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-reference`: sandbox worker mode short-circuits before broker/root setup

## Impact

- Affected code: `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-reference/spec.md`
- No runtime behavior changes
