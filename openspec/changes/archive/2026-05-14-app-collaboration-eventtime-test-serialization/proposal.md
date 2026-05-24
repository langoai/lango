## Why

`internal/app/bridge_collaboration_runtime_test.go` mutates the package-global `eventTime` seam while also marking those tests parallel. That can make repository-wide test results depend on scheduler timing instead of actual behavior.

## What Changes

- Remove parallel execution from the collaboration runtime tests that override `eventTime`
- Keep all assertions and runtime behavior unchanged
- Record the deterministic test requirement in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `test-coverage`: collaboration runtime regressions remain deterministic when package-global time seams are overridden

## Impact

- Affected code: `internal/app/bridge_collaboration_runtime_test.go`
- Affected specs: `openspec/specs/test-coverage/spec.md`
- No runtime behavior change; this is test-suite stabilization
