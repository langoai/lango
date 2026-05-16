## Why

The agent prompt already tells operators that `workflow_run` accepts either `file_path` or `yaml_content` and that the two are mutually exclusive. The implementation still accepted both and silently preferred `file_path`, which leaves the UX ambiguous and the declared contract weaker than the documented one.

## What Changes

- Make `workflow_run` reject calls that provide both `file_path` and `yaml_content`.
- Add regression coverage for the mutual-exclusion error path.
- Sync the automation-agent-tools spec with the explicit mutual-exclusion contract.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `automation-agent-tools`: `workflow_run` now has an explicit mutually-exclusive source-input contract.

## Impact

- Affected code: `internal/workflow/tools.go`
- Affected tests: `internal/workflow/tools_test.go`
- Affected specs: `openspec/specs/automation-agent-tools/spec.md`
