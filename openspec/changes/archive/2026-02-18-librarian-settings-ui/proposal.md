## Why

`LibrarianConfig` is defined in `internal/config/types.go` with 7 fields (Enabled, ObservationThreshold, InquiryCooldownTurns, MaxPendingInquiries, AutoSaveConfidence, Provider, Model) but has no config defaults in the loader and no Settings UI exposure, unlike every other subsystem (Cron, Background, Workflow, etc.).

## What Changes

- Add Librarian default values to `DefaultConfig()` and viper `SetDefault` calls in the config loader.
- Add a "Librarian" menu entry in the Settings TUI between "Workflow Engine" and "Save & Exit".
- Add `NewLibrarianForm()` with all 7 fields (bool, ints, select, text) following existing form patterns.
- Add librarian field cases to `UpdateConfigFromForm()` in state_update.go.
- Add librarian routing in `handleMenuSelection()` in editor.go.

## Capabilities

### New Capabilities

### Modified Capabilities

- `config-system`: Add Librarian default values to DefaultConfig and viper SetDefault bindings.
- `cli-settings`: Add Librarian menu item, form, state update cases, and editor routing.

## Impact

- `internal/config/loader.go` — DefaultConfig + Load function defaults
- `internal/cli/settings/menu.go` — menu category list
- `internal/cli/settings/forms_impl.go` — new form constructor
- `internal/cli/tuicore/state_update.go` — switch cases
- `internal/cli/settings/editor.go` — menu selection handler
