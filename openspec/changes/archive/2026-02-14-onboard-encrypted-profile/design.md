## Context

The `lango onboard` command currently creates a plain-text `lango.json` file via `config.Save()`. Since the bootstrap/configstore refactoring (commit `975c4b0`), all other commands use AES-256-GCM encrypted profiles stored in `~/.lango/lango.db`. The onboard command must be aligned with this new storage model.

Key constraint: `passphrase.Acquire` requires terminal input, and BubbleTea also captures the terminal. These cannot run concurrently — bootstrap must complete before the TUI starts.

## Goals / Non-Goals

**Goals:**
- Onboard saves config via `configstore.Store.Save()` instead of `config.Save()`
- Returning users see their existing profile pre-loaded in the wizard
- New `--profile` flag for named profile creation/editing
- Post-save messaging reflects encrypted storage

**Non-Goals:**
- Changing the TUI wizard forms or navigation
- Modifying bootstrap, configstore, or crypto internals
- Supporting multiple profiles in a single onboard session
- Removing `config.Save()` from the config package (other uses may exist)

## Decisions

### Decision 1: Bootstrap before TUI

Run `bootstrap.Run(Options{})` before `tea.NewProgram()`. This acquires the passphrase, opens the DB, and initializes crypto — all of which need raw terminal access.

**Alternative considered**: Lazy bootstrap after TUI exits. Rejected because the wizard needs to load the existing profile to pre-populate forms.

### Decision 2: Extract save logic from Wizard to runOnboard

The `Wizard.SaveConfig()` method is removed. Instead, `Wizard.Config()` returns the edited config, and `runOnboard()` handles `configstore.Store.Save()`. This keeps the Wizard as a pure TUI model without storage dependencies.

**Alternative considered**: Inject configstore into Wizard. Rejected because it couples the TUI model to infrastructure and complicates testing.

### Decision 3: loadOrDefault pattern

A `loadOrDefault()` helper tries `store.Load()` first. If `ErrProfileNotFound`, it returns `config.DefaultConfig()` and flags `isNew=true`. This determines whether `SetActive()` is called after save.

### Decision 4: PrintNextSteps as package function

`PrintNextSteps` is converted from a Wizard method to a standalone `printNextSteps(cfg, profileName)` function, since it no longer needs wizard internals — just the config and profile name.

## Risks / Trade-offs

- **Double passphrase prompt**: If the user already ran `lango serve` (which also bootstraps), they may need to enter the passphrase again for onboard. → Acceptable; keyfile-based passphrase acquisition can make this transparent.
- **DB left open during TUI**: The `ent.Client` stays open while the user interacts with the wizard. → Minimal risk; SQLite handles this fine with `defer boot.DBClient.Close()`.
- **Existing `lango.json` not cleaned up**: After migration, old JSON files remain. → Out of scope; handled by bootstrap's migration logic.
