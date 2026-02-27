## 1. Remove KeyringConfig from config layer

- [x] 1.1 Delete `KeyringConfig` struct and `Keyring` field from `SecurityConfig` in `internal/config/types.go`
- [x] 1.2 Remove `Keyring: KeyringConfig{Enabled: true}` default in `internal/config/loader.go`
- [x] 1.3 Remove `v.SetDefault("security.keyring.enabled", ...)` in `internal/config/loader.go`
- [x] 1.4 Remove `"keyring": {"enabled": true}` block from `config.json`

## 2. Remove KeyringConfig from settings UI

- [x] 2.1 Delete `NewKeyringForm()` function in `internal/cli/settings/forms_impl.go`
- [x] 2.2 Remove `security_keyring` menu entry in `internal/cli/settings/menu.go`
- [x] 2.3 Remove `case "security_keyring":` block in `internal/cli/settings/editor.go`
- [x] 2.4 Remove `case "keyring_enabled":` block in `internal/cli/tuicore/state_update.go`
- [x] 2.5 Remove "Security Keyring" from help text in `internal/cli/settings/settings.go`

## 3. Add interactive keyring storage prompt

- [x] 3.1 Add `prompt` package import to `internal/bootstrap/bootstrap.go`
- [x] 3.2 Add keyring storage prompt after `passphrase.Acquire()` when source is `SourceInteractive` and keyring provider is available

## 4. Update tests

- [x] 4.1 Delete `TestNewKeyringForm_AllFields` in `internal/cli/settings/forms_impl_test.go`
- [x] 4.2 Remove `"security_keyring"` from `TestNewMenuModel_HasP2PCategories` want list

## 5. Verify

- [x] 5.1 Run `go build ./...` — no compile errors
- [x] 5.2 Run `go test ./internal/config/... ./internal/cli/settings/... ./internal/cli/tuicore/... ./internal/bootstrap/...` — all pass
