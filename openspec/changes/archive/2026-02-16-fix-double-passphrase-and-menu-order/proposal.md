## Why

Security CLI commands (`lango security secrets list`, `set`, `delete`, `status`, `migrate-passphrase`) prompt for the passphrase twice — once during `bootstrap.Run()` and again in `initLocalCrypto()`. This degrades UX by forcing users to enter the same passphrase redundantly. Additionally, the onboarding menu places "Providers" after "Agent", but providers should be configured first since agents depend on them.

## What Changes

- Change security command infrastructure to accept `*bootstrap.Result` instead of `*config.Config`, reusing the already-initialized crypto provider and DB client from bootstrap
- Remove `initLocalCrypto()` function and replace with `secretsStoreFromBoot()` that creates a SecretsStore directly from the bootstrap result
- Refactor `migrateSecrets()` to accept `security.CryptoProvider` interface instead of raw passphrase string, eliminating the current-passphrase prompt
- Reorder onboarding menu to place "Providers" before "Agent"

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `cli-secrets-management`: Security commands now receive full bootstrap result instead of config-only, eliminating redundant passphrase acquisition
- `cli-onboard`: Menu category order changed — Providers now appears first before Agent

## Impact

- `cmd/lango/main.go`: Loader closure returns `*bootstrap.Result` instead of `*config.Config`
- `internal/cli/security/migrate.go`: All function signatures updated, `migrateSecrets` refactored
- `internal/cli/security/secrets.go`: All function signatures updated, uses `secretsStoreFromBoot`
- `internal/cli/security/status.go`: Uses `boot.DBClient` directly instead of opening a new store
- `internal/cli/security/crypto_init.go`: `initLocalCrypto` replaced with `secretsStoreFromBoot`
- `internal/cli/onboard/menu.go`: Category order reordered
