## Why

The `go-keyring` (zalando) dependency provides plain OS keyring access (macOS Keychain, Linux secret-service, Windows DPAPI) which is vulnerable to same-UID attacks. Any process running as the same user can read stored secrets without authentication. Hardware-backed backends (Touch ID via Secure Enclave, TPM 2.0) are immune to this class of attack and are the only keyring backends worth supporting. Removing `go-keyring` reduces attack surface and dependency count.

## What Changes

- **BREAKING**: Remove `OSProvider` struct and `NewOSProvider()` constructor from `internal/keyring/`
- **BREAKING**: Remove `IsAvailable()` function and `Status` struct from `internal/keyring/`
- **BREAKING**: Remove `FallbackProvider` field from `passphrase.Options`
- Remove OSProvider fallback logic from bootstrap passphrase store flow
- Remove OS keyring probe/clear/status logic from CLI `keyring` subcommands
- Remove `github.com/zalando/go-keyring` from `go.mod`
- Remove `security.keyring.enabled` config reference from README
- Delete `openspec/specs/os-keyring/` spec directory
- Delete `openspec/changes/diag-biometric-keychain-error-logging/` change directory
- Update all docs to reflect hardware-only keyring support

**Kept intact:**
- `BiometricProvider` (pure CGO + Apple Security.framework)
- `TPMProvider` (google/go-tpm)
- `DetectSecureProvider()`, `SecurityTier`, `Provider` interface, `KeyChecker` interface

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `passphrase-acquisition`: Remove `FallbackProvider` requirement; acquisition chain is now hardware keyring → keyfile → interactive → stdin
- `bootstrap-lifecycle`: Remove OSProvider fallback wiring and entitlement-to-OSProvider fallback in passphrase store flow

## Impact

- **Code**: `internal/keyring/os_keyring.go` deleted; `internal/bootstrap/bootstrap.go`, `internal/security/passphrase/acquire.go`, `internal/cli/security/keyring.go` modified
- **Dependencies**: `github.com/zalando/go-keyring v0.2.6` removed from go.mod/go.sum
- **Config**: `security.keyring.enabled` setting removed
- **CLI**: `keyring store/clear/status` commands remain but only work with hardware backends
- **Docs**: `docs/security/encryption.md`, `docs/security/index.md`, `docs/cli/security.md`, `README.md` updated
- **Specs**: `openspec/specs/os-keyring/` deleted, `passphrase-acquisition` and `bootstrap-lifecycle` specs updated
