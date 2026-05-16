## Why

The failed completed-turn workbench already uses recovery wording in its starter, hint, placeholder, and footer, but the body lead still says `Pick the next step.`. That leaves one visible copy seam out of sync inside the same recovery state.

## What Changes

- Change the failed completed-turn body lead from `Pick the next step` to `Pick the recovery step`.
- Add regressions for the recovery-lead wording.
- Sync public docs and specs for the recovery-lead copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the failed completed-turn body lead now uses recovery wording consistently.
- `downstream-docs-sync`: public workbench docs describe the recovery-lead wording.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
