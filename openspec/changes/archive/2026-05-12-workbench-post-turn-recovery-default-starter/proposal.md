## Why

The failed completed-turn workbench already talks about recovery steps, recovery prompts, and recovery starters, but its default `Enter` seed still reuses the generic post-turn starter contract. That leaves the recovery loop with a behavior mismatch right where the default action should be most obvious.

## What Changes

- Add a recovery-specific default starter helper for failed completed-turn workbench states.
- Update the workbench page to use that recovery default before the generic completed-turn default.
- Add regressions for the recovery default prompt selection and failed-turn `Enter` seeding.
- Sync public docs and specs for the recovery-oriented `Enter` behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: failed completed-turn `Enter` now seeds a recovery-oriented starter.
- `downstream-docs-sync`: public workbench docs describe the recovery-oriented failed-turn `Enter` behavior.

## Impact

- Affected code: `internal/cli/workbenchstart/prompts.go`, `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbenchstart/prompts_test.go`, `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
