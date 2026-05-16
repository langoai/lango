## Why

The failed completed-turn workbench already switches its body and placeholder to recovery wording, but the footer still says `Type next prompt here`. That leaves one visible mismatch in the same recovery state.

## What Changes

- Change the failed completed-turn footer from `Type next prompt here` to `Type recovery prompt here`.
- Add regressions for the recovery-footer wording.
- Sync public docs and specs for the recovery-footer copy.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the failed completed-turn footer now uses recovery-prompt wording instead of next-prompt wording.
- `downstream-docs-sync`: public workbench docs describe the recovery-footer wording.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
