## Why

The workbench no longer dead-ends incomplete profiles, but a ready profile still lands on a mostly generic empty state. That means the operator has a configured system and no immediate clue what a successful first request should look like.

## What Changes

- Show concrete starter prompts in the workbench empty state when the active profile is ready.
- Keep those starter prompts out of the incomplete-profile path, where setup guidance is more important.
- Update docs and specs so the workbench contract includes the setup-vs-starter split.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Extend the ready-profile workbench empty state with concrete starter prompts.
- `downstream-docs-sync`: Keep workbench docs aligned with the new starter-prompt guidance.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs/specs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`, `openspec/specs/mission-workbench-tui/spec.md`
