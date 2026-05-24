## Why

The post-turn workbench default no longer resets to the original summary starter, but it still uses the same next-step default everywhere. Outside a detected repo, that means the empty workbench can default to `Review recent changes`, which is a poor fit for a generic workspace.

## What Changes

- Make the post-turn empty workbench choose a structure-oriented default outside detected repo context.
- Keep repo-aware workspaces on the stronger next-change default.
- Add regressions and sync docs/specs for the refined post-turn default behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: the post-turn default `Enter` seed becomes context-sensitive, using a structure-oriented starter outside repo context and a next-change starter inside detected workspaces.
- `downstream-docs-sync`: public workbench docs describe the refined post-turn default behavior.

## Impact

- Affected code: `internal/cli/workbenchstart/prompts.go`, `internal/cli/cockpit/pages/missioncontrol_workbench.go`
- Affected tests: `internal/cli/workbenchstart/prompts_test.go`, `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
- Affected specs: `openspec/specs/mission-workbench-tui/spec.md`, `openspec/specs/downstream-docs-sync/spec.md`
