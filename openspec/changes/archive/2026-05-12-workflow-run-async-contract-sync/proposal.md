## Why

The workflow tool implementation already starts workflows asynchronously and returns only a `run_id`, `status`, and launch message. The automation tool spec still described `workflow_run` as if it returned step results immediately, which is a direct contract mismatch.

## What Changes

- Add focused regression coverage for `workflow_run`'s OR-input validation and async return shape.
- Update the automation-agent-tools spec so `workflow_run` is described as an async launcher rather than a synchronous result-returning tool.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `automation-agent-tools`: `workflow_run` now has explicit async launch contract coverage.

## Impact

- Affected tests: `internal/workflow/tools_test.go`
- Affected specs: `openspec/specs/automation-agent-tools/spec.md`
