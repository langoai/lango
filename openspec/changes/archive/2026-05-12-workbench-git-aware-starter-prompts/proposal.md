## Why

The workbench starter prompts already adapt to repository shape, but they still stop short of reflecting the operator's actual live git state. That leaves the “next change” prompt weaker than it should be in an active coding workspace.

## What Changes

- Read branch and dirty-state signals from Git when available for the detected workbench repository.
- Use those signals to sharpen the ready-profile change-review starter prompt.
- Keep the existing repository-aware and Go-aware fallbacks when Git is missing or unavailable.
- Update docs and tests to cover the Git-aware prompt behavior.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Ready-profile starter prompts now reflect live Git branch and dirty-state signals when available.
- `downstream-docs-sync`: Public workbench docs now mention Git-aware quick-start behavior.

## Impact

- Affected code: `internal/cli/workbench/starter_prompts.go`
- Affected tests: `internal/cli/workbench/starter_prompts_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
