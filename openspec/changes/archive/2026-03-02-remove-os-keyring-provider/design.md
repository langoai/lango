## Context

The project currently uses `github.com/zalando/go-keyring` to provide an `OSProvider` that wraps the OS-native keyring (macOS Keychain, Linux secret-service, Windows DPAPI). This provider is used as a fallback when hardware-backed providers (BiometricProvider, TPMProvider) are unavailable or fail with entitlement errors. However, plain OS keyring storage is vulnerable to same-UID attacks — any process running as the same user can read stored secrets without additional authentication.

The hardware-backed providers (`BiometricProvider` via CGO + Apple Security.framework, `TPMProvider` via google/go-tpm) do not use `go-keyring` at all. They are the only consumers that provide meaningful security guarantees.

## Goals / Non-Goals

**Goals:**
- Remove the `go-keyring` dependency and all code that uses it (`OSProvider`, `IsAvailable`, `Status`)
- Simplify the passphrase acquisition chain to: hardware keyring → keyfile → interactive → stdin
- Remove fallback-to-OSProvider logic in bootstrap and CLI
- Update specs, docs, and README to reflect hardware-only keyring support

**Non-Goals:**
- Changing BiometricProvider or TPMProvider behavior
- Removing the `keyring` CLI subcommands (they remain, but only work with hardware backends)
- Adding new passphrase storage mechanisms

## Decisions

### Decision 1: Remove OSProvider entirely rather than deprecate
**Choice**: Delete `os_keyring.go` and all references in a single change.
**Rationale**: OSProvider has no users outside the fallback paths. A deprecation cycle adds complexity for a provider that is actively harmful to security. Clean removal is simpler.
**Alternative considered**: Soft-deprecate with a warning log → rejected because it still leaves the insecure path available.

### Decision 2: Entitlement error → warning instead of OSProvider fallback
**Choice**: When biometric store fails with `ErrEntitlement`, emit a warning and suggest `make codesign` instead of falling back to plain Keychain.
**Rationale**: Falling back to plain Keychain defeats the purpose of hardware-backed security. Users who cannot codesign can use keyfile or interactive prompt.

### Decision 3: Keep `keyring` CLI commands
**Choice**: Retain `keyring store/clear/status` commands but restrict to hardware backends only.
**Rationale**: The commands are still useful for Touch ID and TPM users. Removing them would break existing workflows for users with proper hardware.

### Decision 4: Remove `security.keyring.enabled` config
**Choice**: Remove the config key entirely.
**Rationale**: Hardware keyring detection is automatic via `DetectSecureProvider()`. A config toggle for something that's auto-detected adds confusion.

## Risks / Trade-offs

- **[Users with passphrase in plain OS keyring]** → They will need to re-store using `keyring store` (which now requires hardware), or switch to keyfile/interactive. This is intentional — plain OS keyring was insecure.
- **[macOS users without codesigning]** → They lose the automatic Keychain fallback. Mitigation: clear warning message with `make codesign` tip, plus keyfile/interactive still works.
- **[Windows users]** → Windows had no hardware-backed provider (no TPM provider implemented). They were already using the insecure OSProvider. Now they must use keyfile/interactive. This is a security improvement.
