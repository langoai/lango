## Why

The workbench starter prompt behavior is now richer, but the prompt contract was still split between the workbench package and the Mission Control page package. That leaves unnecessary duplication in the default prompt set and makes future quick-start changes easier to drift.

## What Changes

- Move starter prompt generation into a shared helper package.
- Make the Mission Control page fall back to the shared default prompt contract instead of owning a duplicate constant.
- Preserve all existing user-visible quick-start behavior while reducing architecture drift.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Starter prompt behavior is now produced from one shared helper contract instead of duplicated page-local defaults.

## Impact

- Affected code: `internal/cli/workbenchstart/prompts.go`, `internal/cli/workbench/model.go`, `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbenchstart/prompts_test.go`
