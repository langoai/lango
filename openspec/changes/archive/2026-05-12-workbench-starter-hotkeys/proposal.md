## Why

The workbench already shows starter prompts for ready profiles, but those prompts still rely on the user retyping the suggestion. That keeps the first success path one step too long for the default `lango` entry point.

## What Changes

- Add ready-profile workbench hotkeys `1`, `2`, and `3` to load the starter prompts directly into the composer.
- Update the empty-state copy, composer placeholder, and footer hint so the quick-start path is explicit instead of implied.
- Add regression coverage and align the public docs with the hotkey-driven startup flow.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Ready-profile empty workbench now exposes starter-prompt hotkeys.
- `downstream-docs-sync`: Workbench docs now describe the hotkey-driven starter prompt flow.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench/model_test.go`
- Affected docs: `README.md`, `docs/cli/core.md`, `docs/features/cockpit.md`
