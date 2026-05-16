## Why

The workbench quick-start loop now gets the operator to the first turn quickly, but after the turn completes the result trail was still weaker than it should be because only user submissions and token summaries were reflected into activity.

## What Changes

- Append a short assistant reply summary to the shared activity lane when a turn completes.
- Cover both success and failure result summaries with unit tests.
- Update public docs to mention that the workbench activity lane keeps a visible result summary.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Completed turns now leave an assistant reply summary in the workbench activity lane.
- `downstream-docs-sync`: Public docs now mention assistant reply summaries in activity.

## Impact

- Affected code: `internal/cli/workbench/model.go`, `internal/cli/cockpit/activity_buffer.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
