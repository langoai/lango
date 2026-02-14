## Why

The `lango onboard` command still writes plain-text `lango.json` via `config.Save()`, bypassing the encrypted SQLite profile store (`~/.lango/lango.db`) introduced in the bootstrap/configstore refactoring. This creates an inconsistency where all other commands use encrypted profiles but onboard produces unencrypted files.

## What Changes

- `lango onboard` runs `bootstrap.Run()` before the TUI to initialize DB, crypto, and configstore
- Existing profile is loaded as the initial config for the wizard (supporting re-edit)
- `SaveConfig()` removed from Wizard; saving is done via `configstore.Store.Save()` in `runOnboard()`
- New `--profile` flag allows specifying which profile to create/edit (default: "default")
- New profiles are automatically activated via `configstore.Store.SetActive()`
- Post-save output updated to reference encrypted storage and profile management commands
- Menu text updated from "Write config to file" to "Save encrypted profile"

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `cli-onboard`: Persistence changes from `lango.json` to encrypted profile via configstore. Adds `--profile` flag. Adds `NewWizardWithConfig()` constructor for pre-loading existing config. Removes `SaveConfig()` method. Updates post-save messaging.

## Impact

- **Code**: `internal/cli/onboard/` (onboard.go, wizard.go, state.go, menu.go)
- **Dependencies**: New imports of `bootstrap`, `configstore`, `config` packages in onboard.go
- **User-facing**: `lango onboard` now requires passphrase before TUI starts; no more `lango.json` output
- **Existing profiles**: Returning users see their active profile pre-loaded in the wizard
