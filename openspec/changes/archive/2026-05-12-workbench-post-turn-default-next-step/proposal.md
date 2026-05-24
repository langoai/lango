## Why

The standalone workbench already shortens the loop while a turn is still running, but once that turn finishes the empty workbench falls back to the original starter default. That makes the first-success loop regress from a next-step interaction back to a repo-summary interaction.

## What Changes

- Make the empty ready-profile workbench switch its default `Enter` starter to the next-step starter after at least one turn has completed.
- Add regressions that lock both the copy and key behavior for the post-turn empty state.
- Sync public workbench docs and the mission-workbench spec for the new post-turn default.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the empty workbench default `Enter` prompt changes after a completed turn so the next loop starts from the next-step starter instead of the original summary starter.
- `downstream-docs-sync`: public workbench docs describe the post-turn default starter behavior.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
