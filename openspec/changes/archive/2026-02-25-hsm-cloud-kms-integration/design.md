## Context

The security stack provides `CryptoProvider` (Sign/Encrypt/Decrypt by keyID) with `LocalCryptoProvider` (PBKDF2+AES-256-GCM) and `RPCProvider` (generic delegation). The `CompositeCryptoProvider` already supports primary/fallback with `ConnectionChecker`. `KeyRegistry` tracks keys with `RemoteKeyID` for external mapping. All P0-P2-8 security items are complete.

Cloud KMS SDKs (AWS, GCP, Azure) add heavy transitive dependencies. The project already uses CGO (`mattn/go-sqlite3`), so PKCS#11 (`miekg/pkcs11`) is compatible.

## Goals / Non-Goals

**Goals:**
- Add 4 KMS backends (AWS KMS, GCP KMS, Azure Key Vault, PKCS#11) implementing `CryptoProvider`
- Zero impact on default builds via build tag isolation
- Automatic fallback to local provider when KMS is unavailable
- Retry with exponential backoff for transient KMS errors
- CLI commands for KMS status, testing, and key listing

**Non-Goals:**
- Automatic key rotation orchestration (cloud services handle this transparently for symmetric keys)
- Key creation/provisioning through the CLI (use cloud console/terraform)
- Multi-region KMS failover
- Custom PKCS#11 mechanism configuration beyond ECDSA and AES-GCM

## Decisions

### Build Tag Isolation
Each provider gets two files: implementation (`//go:build kms_aws || kms_all`) and stub (`//go:build !kms_aws && !kms_all`). Stubs return descriptive error messages. This keeps the default binary lean while allowing opt-in compilation.

**Alternative**: Plugin system with `plugin.Open()`. Rejected — Go plugins have platform limitations and complex deployment.

### Provider Selection via signer.provider
Reuse existing `security.signer.provider` config field with new values (`aws-kms`, `gcp-kms`, `azure-kv`, `pkcs11`). KMS-specific config lives under `security.kms.*`.

**Alternative**: Separate `security.kms.provider` field. Rejected — would create ambiguity with `signer.provider`.

### Error Hierarchy with Sentinel Types
KMS-specific sentinel errors (`ErrKMSUnavailable`, `ErrKMSThrottled`, etc.) wrapped in `KMSError` struct. `IsTransient()` helper determines retry eligibility.

**Alternative**: Flat error strings. Rejected — callers need programmatic error classification for retry/fallback decisions.

### Health Checker with Cached Probes
`KMSHealthChecker` probes KMS availability via encrypt/decrypt roundtrip, caching results for 30 seconds. Plugs into existing `CompositeCryptoProvider` via `ConnectionChecker` interface.

**Alternative**: Passive health detection from operation failures. Rejected — would cause latency spikes on first failure after outage.

## Risks / Trade-offs

- [Cloud SDK dependency size] → Mitigated by build tags; only included when explicitly requested
- [KMS API latency (10-100ms per call)] → Mitigated by retry with backoff; fallback to local for availability
- [PKCS#11 CGO requirement] → Already required by `mattn/go-sqlite3`; no additional constraint
- [Stub compilation errors if function signatures drift] → Both stub and impl must match factory signature; compilation catches mismatches immediately
- [Cloud credential misconfiguration] → Clear error messages from SDK default credential chains; config validation catches missing required fields
