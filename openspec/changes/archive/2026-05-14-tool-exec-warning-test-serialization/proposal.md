## Why

`internal/tools/exec/tools_test.go` mutates the package-global `execWarningWriter` while also marking the test parallel. That can make the suite nondeterministic if another future test touches the same seam at the same time.

## What Changes

- Serialize the `execWarningWriter` seam test so it cannot race with sibling tests
- Keep behavioral assertions unchanged
- Record the test-stability contract in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `test-coverage`: exec warning regressions remain deterministic when package-global writer seams are involved

## Impact

- Affected code: `internal/tools/exec/tools_test.go`
- Affected specs: `openspec/specs/test-coverage/spec.md`
- No runtime behavior change; this is test-suite stabilization
