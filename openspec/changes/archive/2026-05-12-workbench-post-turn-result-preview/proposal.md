## Why

The completed-turn workbench already preserves assistant summaries in the activity lane, but the empty body still makes the operator scan that lane to remember what just happened. A compact inline result preview would shorten the next-step loop further.

## What Changes

- Show the latest assistant summary directly in the completed-turn empty workbench body.
- Add regressions for the completed-turn result preview.
- Sync public docs and specs for the new result-preview behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn empty body now includes a compact last-result preview drawn from the assistant activity summary.
- `downstream-docs-sync`: public workbench docs mention the completed-turn result preview.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
