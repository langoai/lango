# Design: KMS/HSM PQ Backends

## Approach

Use existing `CryptoProvider.Encrypt/Decrypt` to wrap/unwrap the Master Key via KMS/HSM. This transitively protects all derived keys (Ed25519, ML-DSA-65, DB key) without requiring direct KMS-backed signing (not feasible — no cloud KMS supports ML-DSA-65).

## Key Decisions

### D1: Reuse `KEKSlotHardware` (already reserved in envelope.go)
No new KEK slot type enum. Repurpose the reserved `"hardware"` type.

### D2: KEKSlot KMS extension fields
Add `KMSProvider` and `KMSKeyID` to `KEKSlot` with `omitempty` for backward compat.

### D3: Envelope version unchanged
Additive `omitempty` fields — no version bump, no migration.

### D4: Environment-based KMS bootstrap config
Chicken-and-egg: KMS config is in encrypted profile, but bootstrap needs KMS to decrypt MK.
Solution: env vars. Provider-specific:
- aws-kms/gcp-kms: `LANGO_KMS_KEY_ID`, `LANGO_KMS_REGION`
- azure-kv: + `LANGO_KMS_AZURE_VAULT_URL`
- pkcs11: `LANGO_KMS_PKCS11_MODULE`, slot/label/PIN

### D5: No CompositeCryptoProvider for MK unwrap
Bare KMS provider only. Local fallback would use wrong key. On failure → fall through to passphrase path.

### D6: No signing backend abstraction
No cloud KMS supports Ed25519/ML-DSA-65. Dead abstraction. Deferred.

### D7: 2-tier KMS slot matching
- Tier 1: exact match (provider + keyID)
- Tier 2: provider-only fallback (handles key rotation)
- No match: `ErrKMSSlotUnavailable` → fall through to passphrase

### D8: Options > env precedence
`Run()` reads env only when `Options.KMSConfig == nil`.

## Architecture

```
Bootstrap KMS path:
  1. Load envelope → detect KEKSlotHardware
  2. Read KMS env vars (or Options) → create bare KMS provider
  3. envelope.UnwrapFromKMS(ctx, kms, providerName, keyID)
     → Tier 1: exact match → KMS.Decrypt(slot.KMSKeyID, wrapped)
     → Tier 2: provider-only match → KMS.Decrypt(slot.KMSKeyID, wrapped)
     → No match: ErrKMSSlotUnavailable
  4. On success: MK unwrapped → skip passphrase → continue as normal
  5. On failure: warn → fall through to mnemonic/passphrase path
```

## Waves

1. **Wave 1**: Envelope KMS Slot — `envelope.go` (AddKMSSlot, UnwrapFromKMS, tests)
2. **Wave 2**: Bootstrap — `kms_env.go`, `phases.go`, `bootstrap.go` (KMS detection, unwrap, fallback)
3. **Wave 3**: CLI + Status — `kms.go` (wrap/detach), `status.go` (KMS protection display)
