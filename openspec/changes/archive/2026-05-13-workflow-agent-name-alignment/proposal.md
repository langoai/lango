## Why

The workflow parser still only recognized older built-in teammate names such as `executor`, `researcher`, and `memory-manager`, while current docs and runtime surfaces describe the built-in registry as `operator`, `navigator`, `vault`, `librarian`, `automator`, `planner`, `chronicler`, and `ontologist`. That made current public workflow examples invalid at parse time.

## What Changes

- Accept current built-in teammate names in workflow YAML validation.
- Keep legacy workflow agent names accepted for backward compatibility.
- Update workflow parser tests to cover both current names and legacy aliases.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `workflow-engine`: workflow YAML now accepts the current built-in teammate registry names, while preserving legacy aliases.

## Impact

- Affected code: `internal/workflow/parser.go`, `internal/workflow/step.go`
- Affected tests: `internal/workflow/parser_test.go`
