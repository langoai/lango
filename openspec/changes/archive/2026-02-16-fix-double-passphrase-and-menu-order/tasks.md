## 1. Security Command Infrastructure

- [x] 1.1 Change `cmd/lango/main.go` security loader from `func() (*config.Config, error)` to `func() (*bootstrap.Result, error)`, remove `defer boot.DBClient.Close()` from closure
- [x] 1.2 Update `NewSecurityCmd` in `migrate.go` to accept `bootLoader func() (*bootstrap.Result, error)` and pass it to all subcommands
- [x] 1.3 Replace `initLocalCrypto` in `crypto_init.go` with `secretsStoreFromBoot(boot *bootstrap.Result)` that creates SecretsStore from boot.Crypto and boot.DBClient

## 2. Secrets Commands

- [x] 2.1 Update `newSecretsListCmd` to use `bootLoader` + `secretsStoreFromBoot`, defer `boot.DBClient.Close()` in RunE
- [x] 2.2 Update `newSecretsSetCmd` to use `bootLoader` + `secretsStoreFromBoot`, defer `boot.DBClient.Close()` in RunE
- [x] 2.3 Update `newSecretsDeleteCmd` to use `bootLoader` + `secretsStoreFromBoot`, defer `boot.DBClient.Close()` in RunE

## 3. Status Command

- [x] 3.1 Update `newStatusCmd` to use `bootLoader`, use `boot.Config` and `boot.DBClient` directly instead of opening a new session store

## 4. Migrate Passphrase Command

- [x] 4.1 Update `newMigratePassphraseCmd` to use `bootLoader`, remove current-passphrase prompt, use `boot.Crypto` as old provider
- [x] 4.2 Refactor `migrateSecrets` to accept `security.CryptoProvider` instead of raw passphrase string

## 5. Onboarding Menu Order

- [x] 5.1 Move "Providers" category to first position in `internal/cli/onboard/menu.go` (before "Agent")

## 6. Verification

- [x] 6.1 Run `go build ./...` and confirm no compilation errors
- [x] 6.2 Run `go test ./...` and confirm all tests pass
