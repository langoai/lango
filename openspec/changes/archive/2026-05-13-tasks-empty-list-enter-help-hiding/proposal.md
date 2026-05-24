## Why

The Tasks page still advertises `Enter` for detail toggling even when the task list is empty, but the runtime ignores `Enter` in that state. That leaves the empty-state help bar advertising an inert key.

## What Changes

- Hide list-mode `Enter` help when the Tasks page has zero task rows.
- Add regressions for the empty-list help surface.
- Update the task-surface spec and cockpit docs to describe the same empty-state contract.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-task-surface`: Tasks list-mode `Enter` help becomes conditional on a selected task row existing.

## Impact

- Affected code: `internal/cli/cockpit/pages/tasks.go`, `internal/cli/cockpit/pages/tasks_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/tui-task-surface/spec.md`
