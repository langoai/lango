## Why

`internal/cli/cockpit/missioncontrol_projector_test.go` overrides the `projector.nowFn` seam in several tests while other tests in the same file run in parallel. That can make suite results depend on scheduler timing rather than actual behavior.

## What Changes

- Remove parallel execution from the mission-control projector tests that override `nowFn`
- Keep assertions and runtime behavior unchanged
- Record the deterministic test requirement in OpenSpec

## Capabilities

### New Capabilities

### Modified Capabilities
- `test-coverage`: mission-control projector regressions remain deterministic when time seams are overridden

## Impact

- Affected code: `internal/cli/cockpit/missioncontrol_projector_test.go`
- Affected specs: `openspec/specs/test-coverage/spec.md`
- No runtime behavior change; this is test-suite stabilization
