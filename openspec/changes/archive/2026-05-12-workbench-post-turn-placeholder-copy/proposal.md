## Why

The completed-turn workbench body and footer already describe a next-step loop, but the empty composer placeholder still uses the original first-run `Press Enter for ...` wording. That leaves one visible copy seam out of sync.

## What Changes

- Change the completed-turn empty composer placeholder to `Next step: press Enter ...`.
- Add regressions for the completed-turn placeholder wording.
- Sync public docs and specs for the refined placeholder copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the completed-turn empty composer placeholder now uses next-step wording.
- `downstream-docs-sync`: public workbench docs describe the completed-turn placeholder as next-step copy.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
