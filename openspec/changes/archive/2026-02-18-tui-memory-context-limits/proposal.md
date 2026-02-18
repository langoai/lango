## Why

The `maxReflectionsInContext` and `maxObservationsInContext` config fields were added to `ObservationalMemoryConfig` but the TUI settings form does not expose them. Users cannot configure these values via `lango settings`.

## What Changes

- Add `om_max_reflections` and `om_max_observations` fields to `ObservationalMemoryForm` in the settings TUI
- Add corresponding state update handlers to persist the values

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `cli-settings`: Add two new fields to the Observational Memory configuration form

## Impact

- **Files**: `internal/cli/settings/forms_impl.go`, `internal/cli/tuicore/state_update.go`
- **Code**: UI-layer only, no core changes
- **APIs**: None
