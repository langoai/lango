## Context

The current bootstrap flow uses `go-keyring` (`OSProvider`) to store/retrieve the master passphrase. On macOS, this uses Keychain; on Linux, D-Bus secret-service. Both allow any process under the same UID to read the stored secret without user interaction, making the passphrase vulnerable to same-UID malicious processes.

The keyring package lives at `internal/keyring/`, bootstrap at `internal/bootstrap/bootstrap.go`, and CLI commands at `internal/cli/security/keyring.go`.

## Goals / Non-Goals

**Goals:**
- Require hardware-backed user presence verification (Touch ID or TPM) before auto-unlocking from keyring
- Gracefully degrade: biometric > TPM > deny (no keyring auto-read)
- Maintain cross-platform compilation via build tags and stubs
- Keep existing keyfile and interactive prompt flows unchanged

**Non-Goals:**
- Windows Hello biometric integration (future work)
- FIDO2/WebAuthn hardware key support
- Removing `go-keyring` / `OSProvider` entirely (still used by CLI commands)
- Encrypting the passphrase at rest beyond what the hardware backend provides

## Decisions

### 1. SecurityTier enum + factory function over config flag
**Decision**: Auto-detect available hardware at runtime via `DetectSecureProvider()`.
**Rationale**: Users shouldn't need to configure security tier manually. The factory probes biometric first, then TPM, then returns nil. This matches the principle of "secure by default."
**Alternative**: Config flag (`security.keyring.tier: biometric|tpm|none`) — rejected because it shifts security responsibility to the user and increases misconfiguration risk.

### 2. CGO for macOS biometric via Security.framework
**Decision**: Use direct CGO calls to `SecAccessControlCreateWithFlags` with `kSecAccessControlBiometryAny`.
**Rationale**: The `go-keyring` library doesn't support biometric ACLs. `kSecAccessControlBiometryAny` ensures any process reading the item triggers Touch ID, which is the core defense against same-UID attacks. No Go-native alternative exists for this Keychain API.
**Alternative**: `osascript` Touch ID prompt — rejected because it's spoofable and doesn't bind to the Keychain item.

### 3. TPM2 seal/unseal via go-tpm
**Decision**: Seal the passphrase under the TPM's Storage Root Key (SRK) and store the blob at `~/.lango/tpm/`.
**Rationale**: TPM-sealed data can only be unsealed by the same TPM chip, providing hardware binding. The SRK is deterministic (same template → same key), so no persistent handle is needed.
**Alternative**: `tpm2-tools` CLI — rejected to avoid external binary dependency.

### 4. Build-tag isolation for platform-specific code
**Decision**: `biometric_darwin.go` (`darwin && cgo`), `tpm_provider.go` (`linux`), with corresponding stubs.
**Rationale**: Ensures clean cross-compilation. Stubs implement Provider interface methods returning sentinel errors, satisfying the type system without runtime code.

### 5. Deny fallback (TierNone) disables keyring auto-read
**Decision**: When no secure hardware is detected, `secureProvider` is nil, effectively skipping keyring in `passphrase.Acquire`.
**Rationale**: Plain OS keyring without hardware protection is the exact vulnerability we're addressing. Denying it forces keyfile or interactive prompt, which are both user-initiated.

### 6. SkipSecureDetection option for testing
**Decision**: Add `SkipSecureDetection bool` to `bootstrap.Options`.
**Rationale**: Tests running on macOS with Touch ID would otherwise trigger biometric prompts or find previously stored passphrases, causing flaky tests. This flag isolates test behavior from host hardware.

## Risks / Trade-offs

- **[Risk] CGO dependency on macOS** → Only affects biometric provider; stub used when CGO disabled. Most macOS Go toolchains have CGO enabled by default.
- **[Risk] Touch ID prompt in non-interactive contexts (SSH, CI)** → `SecItemCopyMatching` returns error quickly when no UI is available; falls through to keyfile.
- **[Risk] TPM device permissions on Linux** → Requires `/dev/tpmrm0` access (typically `tss` group). Doctor command can check this.
- **[Risk] go-tpm API stability** → Using v0.9.x stable API; TPM provider behind build tag limits blast radius.
- **[Trade-off] Two keychain entries possible** → OSProvider and BiometricProvider may create separate entries with same service/account but different ACLs. Clear command cleans both.
