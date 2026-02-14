## 1. Core Store Extension

- [x] 1.1 Add `DeleteReflectionsBySession` method to `internal/memory/store.go`
- [x] 1.2 Add table-driven tests for `DeleteReflectionsBySession` in `internal/memory/store_test.go`

## 2. Memory CLI Commands

- [x] 2.1 Create `internal/cli/memory/memory.go` with `NewMemoryCmd` parent command
- [x] 2.2 Create `internal/cli/memory/list.go` with `newListCmd` and shared `initMemoryStore` helper
- [x] 2.3 Create `internal/cli/memory/status.go` with `newStatusCmd`
- [x] 2.4 Create `internal/cli/memory/clear.go` with `newClearCmd` (confirmation prompt + --force)
- [x] 2.5 Register `memory` command in `cmd/lango/main.go`

## 3. Security Crypto Helper

- [x] 3.1 Create `internal/cli/security/crypto_init.go` with `initLocalCrypto` (passphrase resolution, salt, checksum)

## 4. Secrets CLI Commands

- [x] 4.1 Create `internal/cli/security/secrets.go` with `newSecretsCmd` parent and `list`, `set`, `delete` subcommands
- [x] 4.2 Register `secrets` subcommand in `NewSecurityCmd` in `internal/cli/security/migrate.go`

## 5. Security Status Command

- [x] 5.1 Create `internal/cli/security/status.go` with `newStatusCmd` (key/secret counts, interceptor state)
- [x] 5.2 Register `status` subcommand in `NewSecurityCmd` in `internal/cli/security/migrate.go`

## 6. Doctor Checks

- [x] 6.1 Create `internal/cli/doctor/checks/observational_memory.go` with `ObservationalMemoryCheck`
- [x] 6.2 Create `internal/cli/doctor/checks/output_scanning.go` with `OutputScanningCheck`
- [x] 6.3 Register both checks in `AllChecks()` in `internal/cli/doctor/checks/checks.go`

## 7. Onboard TUI

- [x] 7.1 Add `NewObservationalMemoryForm` to `internal/cli/onboard/forms_impl.go`
- [x] 7.2 Add `observational_memory` category to menu in `internal/cli/onboard/menu.go`
- [x] 7.3 Add `observational_memory` case to `handleMenuSelection` in `internal/cli/onboard/wizard.go`
- [x] 7.4 Add OM field mappings to `UpdateConfigFromForm` in `internal/cli/onboard/state_update.go`

## 8. Verification

- [x] 8.1 Run `go build ./...` and confirm clean compilation
- [x] 8.2 Run `go test ./internal/memory/...` and confirm all tests pass
- [x] 8.3 Run `go test ./internal/cli/...` and confirm all tests pass
- [x] 8.4 Run `go vet ./...` and confirm no warnings
