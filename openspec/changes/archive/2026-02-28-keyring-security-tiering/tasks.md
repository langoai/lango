## 1. Keyring Package Core

- [x] 1.1 Add `ErrBiometricNotAvailable` and `ErrTPMNotAvailable` sentinel errors to `internal/keyring/keyring.go`
- [x] 1.2 Add `SecurityTier` enum with `TierNone`, `TierTPM`, `TierBiometric` and `String()` method to `internal/keyring/keyring.go`
- [x] 1.3 Update `Status` struct with `SecurityTier` field in `internal/keyring/keyring.go`
- [x] 1.4 Create `internal/keyring/tier.go` with `DetectSecureProvider()` factory function

## 2. Biometric Provider (macOS)

- [x] 2.1 Create `internal/keyring/biometric_darwin.go` with CGO Touch ID Keychain implementation (kSecAccessControlBiometryAny)
- [x] 2.2 Create `internal/keyring/biometric_stub.go` with stub returning `ErrBiometricNotAvailable`

## 3. TPM Provider (Linux)

- [x] 3.1 Add `github.com/google/go-tpm` dependency to `go.mod`
- [x] 3.2 Create `internal/keyring/tpm_provider.go` with TPM2 seal/unseal implementation
- [x] 3.3 Create `internal/keyring/tpm_stub.go` with stub returning `ErrTPMNotAvailable`

## 4. OS Keyring Update

- [x] 4.1 Update `IsAvailable()` in `internal/keyring/os_keyring.go` to populate `SecurityTier` in Status

## 5. Bootstrap Integration

- [x] 5.1 Add `SkipSecureDetection` field to `bootstrap.Options`
- [x] 5.2 Replace `OSProvider` with `DetectSecureProvider()` in `bootstrap.Run()`
- [x] 5.3 Update store prompt to show tier label and only trigger when secure provider available
- [x] 5.4 Update existing bootstrap tests with `SkipSecureDetection: true`

## 6. CLI Commands

- [x] 6.1 Update `keyring store` to use `DetectSecureProvider()` and gate on secure provider availability
- [x] 6.2 Update `keyring clear` to clean all backends (OS keyring + secure provider + TPM blob files)
- [x] 6.3 Update `keyring status` to show security tier and per-backend passphrase status

## 7. Tests

- [x] 7.1 Create `internal/keyring/tier_test.go` with SecurityTier and DetectSecureProvider tests
- [x] 7.2 Verify `go build ./...` passes on macOS (biometric_darwin.go + tpm_stub.go)
- [x] 7.3 Verify `go test ./internal/keyring/...` and `go test ./internal/bootstrap/...` pass
