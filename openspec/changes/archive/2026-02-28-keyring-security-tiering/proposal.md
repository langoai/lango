## Why

The current `go-keyring`-based keyring storage allows any process running under the same UID to read the master passphrase without any prompt. This exposes the passphrase to malicious processes. We need hardware-backed user presence verification (biometric/TPM) before allowing automatic keyring unlock, and deny keyring auto-read entirely on systems without secure hardware.

## What Changes

- Add `SecurityTier` enum and `DetectSecureProvider()` factory that probes for the highest available hardware-backed security backend (biometric > TPM > none)
- Add macOS Touch ID Keychain provider (`BiometricProvider`) using `kSecAccessControlBiometryAny` ACL via CGO
- Add Linux TPM 2.0 sealed-blob provider (`TPMProvider`) using `go-tpm` seal/unseal
- Provide build-tag stubs for cross-platform compilation
- Update bootstrap to use `DetectSecureProvider()` instead of plain `OSProvider` — keyring auto-read is disabled when no secure hardware is available (TierNone)
- Update CLI `keyring store/clear/status` commands to reflect security tier and gate store on secure provider availability
- Update `Status` struct with `SecurityTier` field

## Capabilities

### New Capabilities
- `keyring-security-tiering`: Hardware-backed security tier detection, biometric (macOS Touch ID) and TPM 2.0 (Linux) keyring providers, deny-fallback for unsecured environments

### Modified Capabilities
- `os-keyring`: `Status` struct gains `SecurityTier` field; `IsAvailable()` now reports detected tier
- `bootstrap-lifecycle`: Passphrase acquisition switches from plain OSProvider to `DetectSecureProvider()` with `SkipSecureDetection` option
- `passphrase-acquisition`: Keyring provider is now nil when no secure hardware is available (TierNone), effectively disabling keyring auto-read

## Impact

- **Code**: `internal/keyring/` (new files + modified), `internal/bootstrap/bootstrap.go`, `internal/cli/security/keyring.go`
- **Dependencies**: `github.com/google/go-tpm` (new, Linux only via build tags)
- **Build**: CGO required on macOS for biometric provider; stubs used when CGO disabled
- **Tests**: Existing bootstrap tests need `SkipSecureDetection` to avoid Touch ID prompts in CI
