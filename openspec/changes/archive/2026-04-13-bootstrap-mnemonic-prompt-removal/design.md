## Overview

Two-step change: first decouple `recovery restore` from the full bootstrap pipeline so it can operate independently (prerequisite), then remove the mnemonic prompt from bootstrap startup.

## Design Decisions

### D1: Restore uses direct envelope loading, not a lightweight bootstrap

**Decision**: Replace `bootLoader()` call in `newRecoveryRestoreCmd` with direct `security.LoadEnvelopeFile(langoDir)` + `defaultLangoDir()` (already in same package at `status.go:148`).

**Alternatives considered**:
- Add `Options.SkipPassphrase` to bootstrap — too complex, requires handling nil MK in multiple phases.
- Create `RunForRecovery()` — unnecessary abstraction; restore only needs langoDir and envelope.

**Rationale**: Restore does not use DB, config, or crypto provider. It only needs the envelope file and langoDir path. Direct loading is simpler and eliminates the circular dependency (restore needs bootstrap, but bootstrap blocks on passphrase that the user lost).

### D2: LocalCryptoProvider type check replaced by envelope-nil check

**Decision**: The current `boot.Crypto.(*security.LocalCryptoProvider)` assertion is replaced by `envelope == nil` check. `LoadEnvelopeFile` returns `(nil, nil)` when the file doesn't exist, which covers RPC mode (no local envelope).

**Error message**: `"envelope not found — recovery requires local encryption mode"` (preserves semantic equivalence with the old `"local crypto provider only"` message).

### D3: RecoveryMode field preserved as reserved

**Decision**: Keep `State.RecoveryMode` in pipeline.go with updated comment `"reserved for future use"`. The field is no longer set during bootstrap but removing it is unnecessary churn.

## Implementation Steps

### Step 1: Decouple `recovery restore` from bootstrap

**File**: `internal/cli/security/recovery.go`

1. Change `newRecoveryRestoreCmd(bootLoader func() (*bootstrap.Result, error))` → `newRecoveryRestoreCmd()`
2. Replace body:
   - Remove `boot, err := bootLoader()` and `defer boot.DBClient.Close()`
   - Remove `boot.Crypto.(*security.LocalCryptoProvider)` assertion
   - Add: `langoDir := defaultLangoDir()` + empty string check
   - Add: `envelope, err := security.LoadEnvelopeFile(langoDir)` + nil/nil check
   - Replace `boot.LangoDir` references with `langoDir`
   - Keep mnemonic validation, unwrap, ChangePassphraseSlot, keyfile sync, keyring sync unchanged

3. Update `newRecoveryCmd(bootLoader)`:
   - `cmd.AddCommand(newRecoverySetupCmd(bootLoader))` — unchanged
   - `cmd.AddCommand(newRecoveryRestoreCmd())` — no bootLoader arg

### Step 2: Remove mnemonic prompt from bootstrap

**File**: `internal/bootstrap/phases.go`

1. Delete lines 168-191 (recovery path comment + if block)
2. Update doc comment at lines 125-126:
   ```go
   // phaseAcquireCredential attempts KMS unwrap first (if configured),
   // then falls back to the passphrase acquisition chain (keyring, keyfile, interactive, stdin).
   ```
3. Update `phaseUnwrapOrCreateMK` comments at lines 228-231:
   ```go
   //  1. MasterKey already set (by KMS unwrap) — no-op.
   ```
4. Update line 239 comment: `"Already unwrapped (e.g. KMS)."`

**File**: `internal/bootstrap/pipeline.go`

5. Update line 55 comment: `RecoveryMode bool // reserved for future use; no longer set during bootstrap`

## Risks

| Risk | Mitigation |
|------|-----------|
| `defaultLangoDir()` returns empty on `UserHomeDir` failure | Explicit empty-string check before `LoadEnvelopeFile` |
| `recovery setup` still depends on bootLoader (needs current passphrase to unwrap MK) | Only restore is decoupled; setup stays on full bootstrap |
| `LoadEnvelopeFile` returns `(nil, nil)` for missing file, not an error | Explicit `envelope == nil` check after `err == nil` |
