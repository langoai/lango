# Proposal: KMS/HSM PQ Backends

## Why

All derived keys (Ed25519 identity, ML-DSA-65 PQ signing, DB encryption) are protected by the Master Key, which is currently wrapped only with passphrase-derived KEKs (PBKDF2). Production deployments need hardware/cloud-backed key protection to meet enterprise security requirements. No cloud KMS supports ML-DSA-65 natively, so direct KMS-backed PQ signing is not feasible — instead, protecting the MK via KMS transitively protects all derived PQ keys.

## What Changes

- Add KMS KEK slot to `MasterKeyEnvelope` — wrap/unwrap MK using `CryptoProvider.Encrypt/Decrypt` from existing KMS backends (AWS, GCP, Azure, PKCS#11)
- Enable passphraseless bootstrap when KMS is available (env var config, graceful fallback to passphrase)
- Add CLI commands: `lango security kms wrap` (add KMS slot), `lango security kms detach` (remove KMS slot)
- Display KMS protection status in `lango security status`

## Capabilities

### Modified Capabilities
- `cloud-kms` — KMS KEK slot wrapping/unwrapping requirements
- `master-key-envelope` — KEKSlot KMS extension fields, AddKMSSlot, UnwrapFromKMS with 2-tier matching
- `bootstrap-lifecycle` — KMS-based passphraseless bootstrap, env var config precedence
- `cli-security-status` — KMS protection status display

### Downstream Review (no spec changes expected)
- `passphrase-management` — KMS path bypasses passphrase entirely
- `keyring-security-tiering` — KMS slot is orthogonal to keyring tiering
