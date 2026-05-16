## Why

The Tasks page currently advertises `↑/↓` list navigation even when there are zero tasks or only one task, so the help bar exposes inert keys in empty and single-row states. That is inconsistent with the tighter actionability rules already applied to other cockpit pages.

## What Changes

- Hide Tasks list-mode `↑/↓` help when there is no alternative task row to move to.
- Keep list navigation help when two or more tasks exist.
- Add regression coverage and sync the task-surface spec plus cockpit docs with the stricter actionability contract.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-task-surface`: Tasks list-mode navigation help becomes conditional on another task row existing.

## Impact

- Affected code: `internal/cli/cockpit/pages/tasks.go`, `internal/cli/cockpit/pages/tasks_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/tui-task-surface/spec.md`
