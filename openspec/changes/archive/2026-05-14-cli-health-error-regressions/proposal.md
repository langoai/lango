## Why

`lango health` already documents non-200 and timeout failures, but the command-level regressions currently cover only the success path. That leaves two important operator-facing failure contracts verified only by inspection.

## What Changes

- Add a command-level regression for non-200 health responses
- Add a command-level regression for request timeout behavior
- Introduce a small HTTP client seam so timeout behavior is deterministic under test
- Record the failure-path coverage in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `cli-health-check`: failure and timeout paths are regression-covered

## Impact

- Affected code: `cmd/lango/main.go`, `cmd/lango/main_test.go`
- Affected specs: `openspec/specs/cli-health-check/spec.md`
- No runtime behavior changes
