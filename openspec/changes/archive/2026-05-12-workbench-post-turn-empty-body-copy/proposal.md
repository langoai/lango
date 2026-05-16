## Why

The completed-turn workbench now picks a better default starter, but the main empty-state body still leads with `No active missions or pending decisions.`. That reads like a blank dashboard, not like a ready next-step loop after a completed turn.

## What Changes

- Make the completed-turn empty workbench body explicitly say that the last turn completed and the next step is ready.
- Add regressions for the completed-turn body copy.
- Sync public docs and specs for the refined completed-turn body wording.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn empty body now frames the screen as a next-step loop instead of a blank no-missions state.
- `downstream-docs-sync`: public workbench docs describe the completed-turn body as a next-step state.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
