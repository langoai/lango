## Why

The workbench prompt-generation contract was already centralized in a shared helper, but the page wiring still threaded a precomputed prompt slice through `cockpit.Deps`. That kept extra prompt-state plumbing alive even though the page already had the information needed to compute the prompt set itself.

## What Changes

- Remove `WorkbenchStarterPrompts` from `cockpit.Deps`.
- Make the Mission Control workbench page derive starter prompts directly from `workDir` through the shared helper.
- Preserve all user-visible starter prompt behavior while reducing dependency-surface area.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Starter prompt behavior now flows from `workDir` and the shared helper contract rather than from a precomputed prompt slice in TUI dependencies.

## Impact

- Affected code: `internal/cli/cockpit/deps.go`, `internal/cli/cockpit/pages/missioncontrol.go`, `internal/cli/workbench/model.go`
- Affected tests: `internal/cli/workbenchstart`, `internal/cli/workbench`, `internal/cli/cockpit/pages`
