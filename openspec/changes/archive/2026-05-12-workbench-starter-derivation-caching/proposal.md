## Why

The workbench quick-start behavior was already correct, but Mission Control could still re-derive starter prompts during render-time UI updates. Because prompt derivation may inspect the workspace and Git state, recomputing it every render is unnecessary churn on the hottest path.

## What Changes

- Cache the derived starter prompt set and default starter prompt when the Mission Control workbench page is created.
- Keep all existing starter-prompt behavior unchanged.
- Remove render-time re-derivation from the empty-state and footer copy paths.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `mission-workbench-tui`: Starter prompt behavior remains the same, but derivation now happens once per page instance instead of during repeated render work.

## Impact

- Affected code: `internal/cli/cockpit/pages/missioncontrol.go`
- Affected tests: `internal/cli/workbench`, `internal/cli/cockpit/pages`
