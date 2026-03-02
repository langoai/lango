## 1. Remove OSProvider and go-keyring

- [x] 1.1 Delete `internal/keyring/os_keyring.go` (OSProvider, NewOSProvider, IsAvailable, Status, backendName)
- [x] 1.2 Remove `Status` struct from `internal/keyring/keyring.go`
- [x] 1.3 Remove `OSProvider` interface compliance check from `internal/keyring/keyring_test.go`

## 2. Update Bootstrap

- [x] 2.1 Remove `"runtime"` import from `internal/bootstrap/bootstrap.go`
- [x] 2.2 Remove OSProvider fallback provider creation and FallbackProvider wiring
- [x] 2.3 Replace entitlement-to-OSProvider fallback with warning + codesign tip

## 3. Update Passphrase Acquisition

- [x] 3.1 Remove `FallbackProvider` field from `passphrase.Options` struct
- [x] 3.2 Remove step 1b (fallback keyring read) from `Acquire()`

## 4. Update CLI Keyring Commands

- [x] 4.1 Remove `"runtime"` import from `internal/cli/security/keyring.go`
- [x] 4.2 `keyring store`: Replace OSProvider duplicate check with HasKey on secure provider; remove entitlement fallback
- [x] 4.3 `keyring clear`: Remove OS keyring delete section; keep secure provider delete and TPM blob cleanup
- [x] 4.4 `keyring status`: Remove IsAvailable probe and OS keyring passphrase check; use hardware-only status

## 5. Dependency Cleanup

- [x] 5.1 Run `go mod tidy` to remove `github.com/zalando/go-keyring` from go.mod/go.sum
- [x] 5.2 Verify `go build ./...` succeeds
- [x] 5.3 Verify `go test ./internal/keyring/... ./internal/security/passphrase/... ./internal/cli/security/... ./internal/bootstrap/...` passes

## 6. OpenSpec and Documentation

- [x] 6.1 Delete `openspec/specs/os-keyring/` directory
- [x] 6.2 Delete `openspec/changes/diag-biometric-keychain-error-logging/` directory
- [x] 6.3 Update `openspec/specs/passphrase-acquisition/spec.md` — remove FallbackProvider requirement
- [x] 6.4 Update `openspec/specs/bootstrap-lifecycle/spec.md` — remove FallbackProvider wiring and OSProvider fallback requirements
- [x] 6.5 Update `docs/security/encryption.md` — replace OS Keyring section with Hardware Keyring, remove keyring.enabled config
- [x] 6.6 Update `docs/security/index.md` — rename OS Keyring to Hardware Keyring
- [x] 6.7 Update `docs/cli/security.md` — rewrite keyring command docs for hardware-only
- [x] 6.8 Update `README.md` — remove `security.keyring.enabled` row, update OS Keyring section
