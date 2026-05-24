## Why

The Tasks page detail panel always advertises `scroll up` and `scroll down` once detail mode opens, even when the selected task detail fits entirely in the visible panel and there is nothing to scroll. That leaves the help bar advertising inert keys in a common short-detail state.

## What Changes

- Hide `↑/↓` scroll help in Tasks detail mode when the selected task has no scrollable overflow.
- Keep the existing scroll help when detail content actually exceeds the visible panel height.
- Add regression coverage and sync the task-surface spec plus cockpit docs to the stricter actionability rule.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-task-surface`: Tasks detail-mode scroll help becomes conditional on actual scrollability.

## Impact

- Affected code: `internal/cli/cockpit/pages/tasks.go`, `internal/cli/cockpit/pages/tasks_test.go`
- Affected docs/specs: `docs/features/cockpit.md`, `openspec/specs/tui-task-surface/spec.md`
