## Why

The existing `CryptoProvider` stack only supports local (PBKDF2/AES-256-GCM) and generic RPC providers. In production deployments, private keys should never exist in software memory — HSMs and Cloud KMS services provide hardware-backed key management with audit logging, automatic rotation, and compliance guarantees. This is the final item (P2-9) on the security roadmap.

## What Changes

- Add 4 new KMS backend implementations behind build tags: AWS KMS, GCP KMS, Azure Key Vault, PKCS#11
- Add KMS-specific config structs (`KMSConfig`, `AzureKVConfig`, `PKCS11Config`) to `SecurityConfig`
- Add KMS error hierarchy (`ErrKMSUnavailable`, `ErrKMSAccessDenied`, `ErrKMSThrottled`, etc.) with `IsTransient()` helper
- Add retry logic with exponential backoff for transient KMS errors
- Add `KMSHealthChecker` implementing `ConnectionChecker` for KMS liveness probing
- Add `NewKMSProvider()` factory function dispatching to the 4 backends
- Extend `initSecurity()` wiring with KMS provider cases + `CompositeCryptoProvider` fallback
- Add CLI commands: `lango security kms status|test|keys`
- Extend `lango security status` output with KMS fields
- Build tag strategy: `kms_aws`, `kms_gcp`, `kms_azure`, `kms_pkcs11`, `kms_all`

## Capabilities

### New Capabilities
- `cloud-kms`: Cloud KMS and HSM backend integration for CryptoProvider with build-tag isolation, retry logic, health checking, and CLI management

### Modified Capabilities
- `secure-signer`: Extended with KMS provider types in config validation and signer.provider enum
- `key-registry`: KMS keys registered with RemoteKeyID mapped to cloud key ARNs/IDs

## Impact

- **Config**: `security.signer.provider` enum extended with `aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`; new `security.kms.*` block
- **Dependencies**: AWS SDK v2, GCP Cloud KMS, Azure Key Vault SDK, miekg/pkcs11 — all behind build tags, zero impact on default builds
- **CLI**: New `lango security kms` subcommand group
- **Wiring**: `internal/app/wiring.go` initSecurity extended
- **No breaking changes**: `CryptoProvider` interface unchanged
