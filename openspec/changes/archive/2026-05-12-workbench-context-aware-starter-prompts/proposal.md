## Why

The workbench starter prompt hotkeys made the first move shorter, but the prompts themselves were still static. That leaves the default `lango` entry point disconnected from the operator's actual working directory and repository.

## What Changes

- Detect basic workspace context for the standalone workbench from the configured workdir or current directory.
- Upgrade ready-profile starter prompts from static copy to repository-aware prompts when `lango` starts inside a project.
- Use Go-specific structure guidance when the detected workspace includes a `go.mod`.
- Update tests and public docs to reflect the context-aware startup flow.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Starter prompts now adapt to the detected workdir and repository shape.
- `downstream-docs-sync`: Public workbench docs now describe context-aware starter prompts.

## Impact

- Affected code: `internal/cli/workbench/starter_prompts.go`, `internal/cli/workbench/model.go`, `internal/cli/cockpit/pages/missioncontrol.go`, `cmd/lango/main.go`
- Affected tests: `internal/cli/workbench/starter_prompts_test.go`, `internal/cli/workbench/model_test.go`, `cmd/lango/main_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
