## Context

Phase 0-3 of the Security & Crypto roadmap are complete. The handshake protocol (v1.1) uses challenge-response with signature authentication (secp256k1 or Ed25519) but performs no key exchange. Session tokens are HMAC-SHA256 over random bytes. No session encryption key is derived. Message confidentiality relies entirely on libp2p Noise transport.

`cloudflare/circl v1.6.1` is already an indirect dependency and provides `kem/hybrid.X25519MLKEM768()` — a ready-made hybrid KEM combining ML-KEM-768 (FIPS 203, NIST Level 3) with X25519 classical ECDH.

## Goals / Non-Goals

**Goals:**
- Add hybrid PQ KEM (X25519-MLKEM768) to handshake as protocol v1.2
- Derive 32-byte session encryption key via HKDF-SHA256 from hybrid shared secret
- Bind KEM fields to handshake signature transcripts (prevent active tampering)
- Backward compatible with v1.1 peers via omitempty JSON fields
- Ephemeral KEM keypairs per handshake for forward secrecy
- Clean package boundary: circl types do not leak beyond `internal/security`

**Non-Goals:**
- App-level message encryption (Phase 4b)
- PQ signatures (Phase 5)
- KMS/HSM integration (Phase 6)
- Peer capability caching (optimization for Phase 4b+)
- `RequirePQHandshake` enforcement (YAGNI until adoption justifies)

## Decisions

### D1: X25519-MLKEM768 hybrid via circl
Use `circl/kem/hybrid.X25519MLKEM768()`. Shared secret = 64 bytes (ML-KEM 32B || X25519 32B). Public key = 1216 bytes. Ciphertext = 1120 bytes. No custom combiner.

### D2: Protocol v1.2, backward compatible
New `/lango/handshake/1.2.0` protocol ID. KEM fields are `omitempty` — v1.1 peers ignore unknown JSON fields. libp2p multistream-select handles negotiation.

### D3: Session key derivation anchored to signing identity
```
HKDF-SHA256(IKM=hybrid_ss, salt=nil, info="lango-p2p-session-v1:" + signerDID_init + ":" + signerDID_resp, len=32)
```
Uses `initSigner.DID(ctx)` not raw `localDID` parameter. Challenge.SenderDID set from signer DID for wire/HKDF consistency.

### D4: KEM transcript binding
- Challenge canonical payload extended: `nonce || ts || senderDID || kemAlgorithm || kemPublicKey`
- Response canonical payload: `nonce || kemCiphertext` (replaces raw nonce signing)
- Empty KEM fields = same payload as v1.1 (backward compatible)

### D5: KEMDecapsulator closure pattern
`GenerateEphemeralKEM()` returns `(pubKeyBytes, KEMDecapsulator, error)`. `KEMDecapsulator` is `func(ct []byte) (ss []byte, err error)` — a closure capturing the circl `kem.PrivateKey`. Handshake package never imports circl.

### D6: Initiator protocol fallback
Single initiator call site (`tools_p2p.go:128`). `PreferredProtocols(kemEnabled)` returns `[v1.2, v1.1, v1.0]` or `[v1.1, v1.0]`. libp2p `NewStream` tries in order.

### D7: EncryptionKey zeroing policy
`security.ZeroBytes()` called at: `Create()` overwrite, `Remove()`, `Invalidate()`, `InvalidateAll()`, `InvalidateByCondition()`, `Cleanup()`. KEM private key inside closure is GC-managed (residual risk accepted).

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| JSON payload size (+3KB base64 per handshake) | Low | One-time per peer |
| KEM private key in GC-managed closure | Low residual | Shared secret explicitly zeroed; Phase 6 KMS/HSM is the long-term answer |
| Shared secret logging | High if violated | Only pass to `DeriveSessionKey`, immediately zero |
| v1.1 backward compat regression | Medium | Existing tests must pass unchanged |
| Session key derived but not used for encryption | Accepted | Phase 4b adds message encryption; `KEMUsed` flag enables policy now |
