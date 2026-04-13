## Why

Every interactive startup of `lango` prompts "Recovery mnemonic slot detected. Recover with mnemonic? [y/N]:" when the envelope contains a mnemonic KEK slot. This interrupts normal usage on every launch. The `lango security recovery restore` command already provides dedicated mnemonic recovery, making the bootstrap prompt redundant and harmful to UX. Additionally, `recovery restore` currently depends on the full bootstrap pipeline, which means a user who lost their passphrase cannot even reach the restore command — the exact scenario recovery is designed for.

## What Changes

- **Decouple `recovery restore` from full bootstrap**: Replace `bootLoader()` dependency with direct envelope loading via `security.LoadEnvelopeFile()`, using the existing `defaultLangoDir()` helper from the same package.
- **Remove mnemonic prompt from bootstrap**: Delete the mnemonic recovery prompt block in `phaseAcquireCredential()` (phases.go:168-191). Bootstrap will proceed directly from KMS attempt to passphrase acquisition.
- **Update comments and state fields**: Clean up mnemonic-related comments in `phaseUnwrapOrCreateMK()` and mark `RecoveryMode` as reserved in `pipeline.go`.
- **Update specs**: Align recovery-mnemonic, passphrase-acquisition, and bootstrap-lifecycle specs with the new behavior.

## Capabilities

### New Capabilities

_(none)_

### Modified Capabilities

- `recovery-mnemonic`: Bootstrap no longer offers mnemonic choice; recovery is exclusively via `lango security recovery restore` which loads envelope directly without full bootstrap.
- `passphrase-acquisition`: Remove the recovery credential choice during bootstrap Phase 4. Passphrase acquisition always follows the standard priority chain.
- `bootstrap-lifecycle`: KMS fallback description changes from `(mnemonic → passphrase)` to `passphrase` only.

## Impact

- **`internal/cli/security/recovery.go`**: `newRecoveryRestoreCmd()` loses `bootLoader` parameter; loads envelope directly.
- **`internal/cli/security/recovery.go`**: `newRecoveryCmd()` wiring updated (setup keeps bootLoader, restore independent).
- **`internal/bootstrap/phases.go`**: Lines 168-191 deleted, comments at lines 125-126, 228-231, 239 updated.
- **`internal/bootstrap/pipeline.go`**: `RecoveryMode` comment updated.
- **`internal/cli/security/migrate.go`**: No signature change needed (NewSecurityCmd stays the same).
- **`internal/cli/security/security_test.go`**: No change needed (public API unchanged).
- **Test additions**: `recovery_test.go` for direct envelope load path; `bootstrap_envelope_test.go` for mnemonic-slot-present behavioral test.
