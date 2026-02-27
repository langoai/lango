## Why

The `security.keyring.enabled` config flag is redundant — `keyring.IsAvailable()` already provides runtime auto-detection of OS keyring availability. Additionally, after interactive passphrase entry there is no UX prompt to store the passphrase in the keyring, forcing users to manually run `lango security keyring store`.

## What Changes

- **Remove** `security.keyring.enabled` config flag, `KeyringConfig` struct, all related defaults, TUI form, menu entry, and state update handler.
- **Add** an interactive prompt in `bootstrap.go` after `passphrase.Acquire()` that offers to store the passphrase in the OS keyring when the source is interactive and the keyring is available.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `os-keyring`: Remove the `security.keyring.enabled` config flag (runtime auto-detection is sufficient) and add automatic keyring storage prompt after interactive passphrase entry.

## Impact

- `internal/config/types.go` — Remove `KeyringConfig` struct and `Keyring` field from `SecurityConfig`
- `internal/config/loader.go` — Remove keyring defaults
- `internal/cli/settings/` — Remove `NewKeyringForm()`, menu entry, editor case
- `internal/cli/tuicore/state_update.go` — Remove `keyring_enabled` case
- `internal/bootstrap/bootstrap.go` — Add keyring storage prompt after interactive passphrase acquisition
- `config.json` — Remove `keyring` block from security section
- Tests updated to remove keyring form test and menu assertion
