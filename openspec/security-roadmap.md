# P2P Security Hardening Roadmap

## Context

The P2P node key (`~/.lango/p2p/node.key`) is stored as plaintext binary protected only by file permissions (`0600`). Meanwhile, wallet keys are properly encrypted in `SecretsStore` (AES-256-GCM), creating an **architectural inconsistency**. Additionally, handshake signature verification is incomplete, session invalidation is absent, and tool execution lacks process isolation.

This roadmap addresses security hardening in three phases: P0 (immediate) → P1 (medium-term) → P2 (long-term).

---

## Current Security Posture

| Area | Grade | Notes |
|------|-------|-------|
| Crypto primitives | **A** | AES-256-GCM, PBKDF2 100K iter, HMAC-SHA256 |
| Wallet key management | **A-** | Encrypted storage + memory zeroing |
| P2P authentication | **A** | Signed challenge, nonce replay protection, timestamp validation, dual protocol versioning |
| P2P node key management | **A** | Encrypted via SecretsStore (P0-1) |
| ZK proofs | **B+** | Full test suite, timestamp freshness, capability binding fix, structured attestation, SRS file support |
| Session management | **A-** | TTL + explicit invalidation + security events (P1-6) |
| Response sanitization | **A-** | Owner Shield + sensitive field removal |
| Execution isolation | **A-** | Subprocess + Container sandbox with Docker (P1-5, P2-8) |
| DB encryption | **A** | SQLCipher transparent encryption (P2-7) |
| OS Keyring | **A** | macOS/Linux/Windows keyring integration (P1-4) |
| HSM/Cloud KMS | **A** | AWS/GCP/Azure/PKCS#11 with build-tag isolation (P2-9) |

---

## P0: Critical (Immediate)

### P0-1: Migrate P2P Node Key to SecretsStore

**Problem:** `internal/p2p/node.go:200-245` stores Ed25519 private key as plain binary at `~/.lango/p2p/node.key` with only `0600` permissions. This is inconsistent with wallet keys which are encrypted in `SecretsStore`.

**Files to modify:**
- `internal/p2p/node.go` — Refactor `loadOrGenerateKey(keyDir, secrets)` to use SecretsStore
- `internal/app/wiring.go` — Add `*security.SecretsStore` parameter to `initP2P`
- `internal/cli/p2p/p2p.go` — Build SecretsStore in `initP2PDeps` and pass to `NewNode`
- `internal/cli/p2p/identity.go` — Replace `keyDir` output with `keyStorage` info

**Design:**
1. Change signature: `loadOrGenerateKey(keyDir string, secrets *security.SecretsStore)`
2. Priority: SecretsStore → Legacy file → Generate new key
3. Auto-migration: When legacy `node.key` found, store in SecretsStore then delete plaintext file
4. Fallback: When `secrets == nil`, retain file-based storage for backward compatibility
5. Apply `zeroBytes()` pattern for immediate memory cleanup (from `internal/wallet/local_wallet.go:153`)
6. Migration failure: warn log only, don't block startup (retry on next restart)

**Reuse:** `SecretsStore.Store()/Get()` (`internal/security/secrets_store.go`), `zeroBytes()` pattern (`internal/wallet/local_wallet.go:153`)

**Key constant:** `nodeKeySecret = "p2p.node.privatekey"` in SecretsStore

### P0-2: Complete Handshake Signature Verification

**Problem:** `internal/p2p/handshake/handshake.go:293-300` accepts any non-empty signature as valid.

**Current vulnerable code:**
```go
if len(resp.Signature) > 0 {
    // For now, we accept signatures as valid if they are non-empty.
    // Full ECDSA recovery verification will be added in integration.
    return nil
}
```

**Files to modify:**
- `internal/p2p/handshake/handshake.go:293-300` — Replace stub with real verification

**Design:**
1. Hash message: `ethcrypto.Keccak256(nonce)` (matches wallet `SignMessage` pattern)
2. Recover public key: `ethcrypto.SigToPub(hash, signature)` (secp256k1 ECDSA recovery)
3. Compare: `ethcrypto.CompressPubkey(recovered)` vs `resp.PublicKey`
4. Replace nonce comparison with `hmac.Equal()` (constant-time) to prevent timing attacks
5. Validate signature length == 65 bytes (R32 + S32 + V1)

**Imports to add:** `ethcrypto "github.com/ethereum/go-ethereum/crypto"`, `"bytes"`, `"crypto/hmac"`

**Reuse:** `go-ethereum/crypto` (already used by wallet)

### P0-3: Clean Up KeyDir Config Exposure

**Files to modify:**
- `internal/config/types.go` — Add `omitempty` to `KeyDir`, mark as deprecated
- `internal/config/loader.go` — Add `nodeKeyName` default value

---

## P1: Medium-term

### P1-4: OS Keyring Integration

**Rationale:** Master passphrase currently acquired from keyfile (disk plaintext) or interactive input. OS keyring (macOS Keychain / Linux secret-service / Windows DPAPI) provides hardware-backed protection without leaving keyfiles on disk.

**Files to create/modify:**
- New: `internal/keyring/keyring.go`, `internal/keyring/os_keyring.go`
- `internal/passphrase/acquire.go` — Add keyring source (priority: keyring → keyfile → interactive → stdin)
- `internal/bootstrap/bootstrap.go` — Keyring integration
- `internal/config/types.go` — Add `KeyringConfig`

**Design:**
- Library: `github.com/zalando/go-keyring` (cross-platform)
- Service name: `"lango"`, Key: `"master-passphrase"`
- No `CryptoProvider` interface changes needed
- CLI: `lango security keyring store/clear/status`
- Graceful fallback when keyring daemon unavailable (Linux CI environments)

**Dependencies:** None (independent)
**Complexity:** Medium (2-3 days)
**Risk:** Low — existing paths (keyfile/interactive) retained as fallback

### P1-5: Tool Execution Process Isolation

**Rationale:** `handler.go:236` executes `h.executor(ctx, toolName, params)` in-process. Malicious tool invoked by remote peer can access process memory (passphrases, private keys, session tokens).

**Files to create/modify:**
- New: `internal/sandbox/executor.go`, `internal/sandbox/subprocess.go`
- `internal/p2p/protocol/handler.go` — Route remote peer requests through `SubprocessExecutor`
- `internal/config/types.go` — Add `ToolIsolationConfig`

**Design:**
- `Executor` interface: `InProcessExecutor` (local) + `SubprocessExecutor` (remote peer)
- JSON stdin/stdout communication protocol
- `context.WithTimeout` + `cmd.Process.Kill()` for forced termination
- Resource limits Phase 1: timeout only; Phase 2: rlimit (Linux)
- Config: `p2p.toolIsolation.enabled`, `timeoutPerTool`, `maxMemoryMb`

**Dependencies:** None (but prerequisite for P2-8)
**Complexity:** High (4-5 days)
**Risk:** Medium — subprocess overhead (tens of ms latency)

### P1-6: Session Explicit Invalidation

**Rationale:** `SessionStore` (`internal/p2p/handshake/session.go`) only supports TTL-based expiration. No explicit logout, security-event-based revocation, or session listing.

**Files to create/modify:**
- `internal/p2p/handshake/session.go` — Add `Invalidate()`, `InvalidateAll()`, `InvalidateByCondition()`
- New: `internal/p2p/handshake/security_events.go` — Auto-invalidation event handler
- `internal/p2p/protocol/handler.go` — Enhanced session validation, consecutive failure tracking
- `internal/p2p/reputation/store.go` — Reputation drop → session invalidation callback

**Design:**
- `InvalidationReason` type: `logout`, `reputation_drop`, `repeated_failures`, `manual_revoke`, `security_event`
- Auto-invalidate when reputation drops below `minTrustScore`
- Auto-invalidate after N consecutive tool execution failures
- CLI: `lango p2p session revoke/list/revoke-all`

**Dependencies:** None (independent)
**Complexity:** Medium (2-3 days)
**Risk:** Low — additive feature, existing TTL mechanism unaffected

---

## P2: Long-term

### P2-7: SQLCipher DB Transparent Encryption ✅ COMPLETED (2026-02-25)

**Status:** Implemented. Kept `mattn/go-sqlite3` driver (sqlite-vec compatibility) with PRAGMA-based SQLCipher encryption. System `libsqlcipher-dev` required at build time.

**Implementation:**
- `internal/bootstrap/bootstrap.go` — Restructured: detect encryption → acquire passphrase first → `PRAGMA key` + `PRAGMA cipher_page_size` after `sql.Open`
- `internal/dbmigrate/migrate.go` — `MigrateToEncrypted()`, `DecryptToPlaintext()`, `IsEncrypted()`, `secureDeleteFile()`
- `internal/cli/security/db_migrate.go` — `lango security db-migrate`, `lango security db-decrypt` with `--force`
- `internal/cli/security/status.go` — DB encryption status display
- `internal/config/types.go` — `DBEncryptionConfig{Enabled, CipherPageSize}`
- Config: `security.dbEncryption.{enabled,cipherPageSize}`
- Spec: `openspec/specs/db-encryption/spec.md`

### P2-8: Container-based Tool Execution Sandbox ✅ COMPLETED (2026-02-25)

**Status:** Implemented. Docker Go SDK-based container isolation with NativeRuntime fallback.

**Implementation:**
- `internal/sandbox/container_runtime.go` — `ContainerRuntime` interface, error types
- `internal/sandbox/docker_runtime.go` — Docker SDK implementation (full lifecycle, OOM detection, label-based cleanup)
- `internal/sandbox/native_runtime.go` — SubprocessExecutor wrapper as fallback
- `internal/sandbox/gvisor_runtime.go` — Stub (future implementation)
- `internal/sandbox/container_executor.go` — Runtime probe chain (Docker → gVisor → Native)
- `internal/sandbox/container_pool.go` — Optional pre-warmed container pool
- `internal/cli/p2p/sandbox.go` — `lango p2p sandbox status|test|cleanup`
- `internal/app/app.go` — Container sandbox wiring with subprocess fallback
- `build/sandbox/Dockerfile` — Minimal sandbox image
- Config: `p2p.toolIsolation.container.{enabled,runtime,image,networkMode,readOnlyRootfs,cpuQuotaUs,poolSize,poolIdleTimeout}`
- Spec: `openspec/specs/container-sandbox/spec.md`

### P2-9: HSM / Cloud KMS Integration ✅ COMPLETED (2026-02-25)

**Status:** Implemented. Build-tag based isolation ensures Cloud SDK dependencies are only included when explicitly opted in. Four KMS backends available: AWS KMS, GCP KMS, Azure Key Vault, PKCS#11.

**Implementation:**
- `internal/security/kms_factory.go` — `NewKMSProvider()` factory dispatching to 4 backends
- `internal/security/kms_retry.go` — Exponential backoff with transient error detection
- `internal/security/kms_checker.go` — `KMSHealthChecker` implementing `ConnectionChecker` with 30s probe cache
- `internal/security/errors.go` — KMS sentinel errors (`ErrKMSUnavailable`, `ErrKMSAccessDenied`, `ErrKMSThrottled`, etc.) + `KMSError` type + `IsTransient()` helper
- `internal/security/aws_kms_provider.go` — AWS SDK v2 implementation (ECDSA_SHA_256 signing, SYMMETRIC_DEFAULT encrypt/decrypt)
- `internal/security/gcp_kms_provider.go` — GCP Cloud KMS implementation (AsymmetricSign SHA-256, symmetric encrypt/decrypt)
- `internal/security/azure_kv_provider.go` — Azure Key Vault implementation (ES256 signing, RSA-OAEP encrypt/decrypt)
- `internal/security/pkcs11_provider.go` — PKCS#11 HSM implementation (CKM_ECDSA signing, CKM_AES_GCM encrypt/decrypt with IV prepend)
- `internal/security/*_stub.go` — Stub files for uncompiled providers (4 files)
- `internal/security/kms_all.go` — Build tag grouping (`kms_all`)
- `internal/config/types.go` — `KMSConfig`, `AzureKVConfig`, `PKCS11Config` structs
- `internal/config/loader.go` — KMS defaults + config validation for each provider
- `internal/app/wiring.go` — `initSecurity()` KMS case with `CompositeCryptoProvider` fallback
- `internal/cli/security/kms.go` — `lango security kms status|test|keys`
- `internal/cli/security/status.go` — KMS fields in status output
- Build tags: `kms_aws`, `kms_gcp`, `kms_azure`, `kms_pkcs11`, `kms_all`
- Config: `security.signer.provider: "aws-kms"|"gcp-kms"|"azure-kv"|"pkcs11"`, `security.kms.*`

### P2-10: Signed Challenge & Nonce Replay Protection ✅ COMPLETED (2026-02-25)

**Status:** Implemented. Challenges now carry ECDSA signature over canonical payload (nonce || timestamp || senderDID). Dual protocol versioning (v1.0 legacy + v1.1 signed).

**Implementation:**
- `internal/p2p/handshake/handshake.go` — Challenge struct extended with PublicKey/Signature; Initiate() signs challenges; HandleIncoming() validates timestamp, nonce replay, and signature
- `internal/p2p/handshake/nonce_cache.go` — TTL-based nonce deduplication with periodic cleanup goroutine
- Protocol versioning: `ProtocolID="/lango/handshake/1.0.0"` (legacy), `ProtocolIDv11="/lango/handshake/1.1.0"` (signed)
- Config: `p2p.requireSignedChallenge` (default: false for backward compat)
- Timestamp window: 5 min past + 30s future grace

### P2-11: ZK Circuit Hardening ✅ COMPLETED (2026-02-25)

**Status:** Implemented. Full test coverage for all 4 circuits, attestation timestamp freshness, capability binding fix, structured attestation data, SRS production path.

**Implementation:**
- `internal/zkp/circuits/circuits_test.go` — 15 test cases across 4 circuits (gnark test framework, BN254 curve, both plonk and groth16)
- `internal/zkp/zkp_test.go` — 6 ProverService integration tests (compile, prove, verify, tamper detection, idempotent compile, uncompiled error)
- `internal/zkp/circuits/attestation.go` — MinTimestamp/MaxTimestamp public inputs with range assertions
- `internal/zkp/circuits/capability.go` — AgentTestBinding public field properly constrained (was discarded)
- `internal/zkp/zkp.go` — SRS file loading support (SRSMode "unsafe"|"file")
- `internal/p2p/protocol/messages.go` — AttestationData struct with proof, public inputs, circuit ID, scheme
- `internal/p2p/firewall/firewall.go` — AttestationResult struct, ZKAttestFunc returns structured data
- `internal/p2p/protocol/handler.go` — Constructs AttestationData in both tool invoke paths
- `internal/p2p/protocol/remote_agent.go` — ZKAttestVerifyFunc callback for attestation verification
- Config: `p2p.zkp.srsMode`, `p2p.zkp.srsPath`, `p2p.zkp.maxCredentialAge`

### P2-12: Credential Revocation ✅ COMPLETED (2026-02-25)

**Status:** Implemented. Gossip discovery now checks credential max age and revoked DIDs.

**Implementation:**
- `internal/p2p/discovery/gossip.go` — revokedDIDs map, RevokeDID()/IsRevoked() methods, maxCredentialAge validation, SetMaxCredentialAge() setter
- Credential rejection: expired (ExpiresAt), stale (IssuedAt + maxCredentialAge), revoked (IsRevoked)

---

## P3: Future (post-hardening)

| Item | Area | Description |
|------|------|-------------|
| P3-1 | Authentication | Mutual TLS certificate pinning for bootstrap peers |
| P3-2 | ZK Proofs | Recursive proof composition (aggregate multiple attestations) |
| P3-3 | ZK Proofs | Production SRS ceremony (replace unsafe KZG setup) |
| P3-4 | Credentials | DID credential rotation protocol |
| P3-5 | Credentials | Verifiable Credential (W3C VC) integration |
| P3-6 | Monitoring | Security audit logging with tamper-evident storage |
| P3-7 | Network | Tor/I2P transport layer support |

---

## Dependency Graph & Execution Order

```
P0-1 (Node key SecretsStore) ──┐
P0-2 (Signature verification)  ├── Immediate (1 week)
P0-3 (KeyDir cleanup)         ──┘

P1-4 (OS Keyring)  ────────── Independent ──→ P2-7 synergy
P1-5 (Process isolation) ───── Independent ──→ P2-8 prerequisite
P1-6 (Session invalidation) ── Independent
                                                (All P1 parallelizable, 2-3 weeks)

P2-7 (SQLCipher)   ────────── After P1-4
P2-8 (Container)   ────────── After P1-5
P2-9 (HSM/KMS)    ────────── Independent
P2-10 (Signed Challenge) ──── After P0-2
P2-11 (ZK Hardening) ─────── Independent
P2-12 (Credential Revocation) Independent
                                                (P2: completed)

P3-1..P3-7 ────────────────── Future
```

## Risk Matrix

| Item | Impact | Complexity | Failure Risk | Compat Risk | Overall |
|------|--------|-----------|-------------|-------------|---------|
| P0-1 Node key migration | High | Medium | Low | Low | **Low** |
| P0-2 Signature verification | High | Low | Low | Low | **Low** |
| P0-3 KeyDir cleanup | Low | Low | Low | Low | **Low** |
| P1-4 OS Keyring | Medium | Medium | Low | Low | **Low** |
| P1-5 Process isolation | High | High | Medium | Low | **Medium** |
| P1-6 Session invalidation | Medium | Medium | Low | Low | **Low** |
| P2-7 SQLCipher | High | High | Medium | Medium | **Medium-High** |
| P2-8 Container sandbox | High | Very High | High | Low | **High** |
| P2-9 HSM/Cloud KMS | High | Very High | Medium | Low | **Medium** |
