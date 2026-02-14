## 1. Passphrase Acquisition Package

- [x] 1.1 Create `internal/passphrase/acquire.go` with `Acquire()` priority chain (keyfile → interactive → stdin)
- [x] 1.2 Create `internal/passphrase/keyfile.go` with `ReadKeyfile`, `WriteKeyfile`, `ValidatePermissions`
- [x] 1.3 Create `internal/passphrase/stdin.go` with `ReadStdinPipe`
- [x] 1.4 Write unit tests for all passphrase sources and edge cases

## 2. ConfigProfile Ent Schema

- [x] 2.1 Create `internal/ent/schema/config_profile.go` with UUID, name, encrypted_data, active, version, timestamps
- [x] 2.2 Run `go generate ./internal/ent/...` to generate ORM code
- [x] 2.3 Verify generated code compiles (`go build ./internal/ent/...`)

## 3. Configstore Package

- [x] 3.1 Create `internal/configstore/types.go` with `ProfileInfo` struct
- [x] 3.2 Create `internal/configstore/store.go` with `Store` (Save/Load/LoadActive/SetActive/List/Delete/Exists)
- [x] 3.3 Create `internal/configstore/migrate.go` with `MigrateFromJSON`
- [x] 3.4 Write unit tests for Save/Load round-trip, List, Delete, SetActive, wrong passphrase

## 4. Bootstrap Package

- [x] 4.1 Create `internal/bootstrap/bootstrap.go` with `Run()` function and `Result`/`Options` types
- [x] 4.2 Implement `openDatabase()` helper (SQLite open + ent schema migration)
- [x] 4.3 Implement `loadSecurityState()` helper (salt/checksum detection)
- [x] 4.4 Implement `handleNoProfile()` helper (auto-migration + default creation)
- [x] 4.5 Verify bootstrap builds and integrates with passphrase + configstore

## 5. Session Store Update

- [x] 5.1 Add `NewEntStoreWithClient()` constructor to `internal/session/ent_store.go`
- [x] 5.2 Update `Close()` to be a no-op when `db` is nil (externally managed client)

## 6. App Layer Modification

- [x] 6.1 Change `app.New(*config.Config)` to `app.New(*bootstrap.Result)` in `internal/app/app.go`
- [x] 6.2 Update `initSessionStore()` in `wiring.go` to reuse `boot.DBClient` via `NewEntStoreWithClient`
- [x] 6.3 Update `initSecurity()` in `wiring.go` to reuse `boot.Crypto` for local provider
- [x] 6.4 Remove `LANGO_PASSPHRASE` env var references from `wiring.go`
- [x] 6.5 Add `BlockedPaths` (`~/.lango/`) to filesystem tool config
- [x] 6.6 Add `registerConfigSecrets()` to register provider/channel secrets with SecretScanner
- [x] 6.7 Update `app_test.go` to use `*bootstrap.Result` signature

## 7. CLI Updates

- [x] 7.1 Update `cmd/lango/main.go` serve command to use `bootstrap.Run()`
- [x] 7.2 Add `config list` subcommand with tabwriter table output
- [x] 7.3 Add `config create <name>` subcommand with duplicate check
- [x] 7.4 Add `config use <name>` subcommand for profile switching
- [x] 7.5 Add `config delete <name>` subcommand with confirmation + `--force` flag
- [x] 7.6 Add `config import <file>` subcommand with `--profile` flag
- [x] 7.7 Add `config export <name>` subcommand with stderr security warning
- [x] 7.8 Update `config validate` to work with bootstrap-loaded active profile
- [x] 7.9 Update `clisecurity` and `climemory` config loader callbacks to use bootstrap

## 8. Agent Security Hardening

- [x] 8.1 Add `BlockedPaths` field to `filesystem.Config` and blocked path check in `validatePath()`
- [x] 8.2 Add `LANGO_PASSPHRASE` to exec tool environment variable blacklist
- [x] 8.3 Write tests for BlockedPaths enforcement and env var filtering

## 9. Deprecated Code Cleanup

- [x] 9.1 Remove `SecurityConfig.Passphrase` field from `internal/config/types.go`
- [x] 9.2 Remove passphrase substitution from `config/loader.go`, simplify `Save()`
- [x] 9.3 Update `cli/security/crypto_init.go` to use passphrase package (remove env var/config references)
- [x] 9.4 Update `cli/security/secrets.go` call sites for new `initLocalCrypto` signature
- [x] 9.5 Remove LANGO_PASSPHRASE and deprecated passphrase checks from `doctor/checks/security.go`
- [x] 9.6 Remove passphrase form field from `cli/onboard/forms_impl.go` and `state_update.go`
