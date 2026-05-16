## Why

The workbench already let `Enter` seed a starter prompt, but it always chose the first prompt in the list. In a dirty repository, the most relevant first move is usually change review rather than repository summary.

## What Changes

- Make the default `Enter` quick-start prompt follow repository state.
- Use the changed-review starter prompt as the default when the detected workspace is dirty.
- Preserve summary-first defaulting for clean or non-repository workspaces.
- Update tests and docs to reflect that `Enter` seeds a context-aware default prompt, not a hard-coded one.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Default `Enter` quick-start prompt now follows repository state instead of always choosing the summary prompt.
- `downstream-docs-sync`: Public docs now describe the default prompt as context-aware.

## Impact

- Affected code: `internal/cli/workbenchstart/prompts.go`, `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbenchstart/prompts_test.go`, `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
