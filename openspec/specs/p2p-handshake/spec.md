## Purpose

Capability spec for p2p-handshake. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Challenge-Response Mutual Authentication

The `Handshaker` SHALL accept a `Signer` interface (methods `SignMessage(ctx, message) ([]byte, error)`, `PublicKey(ctx) ([]byte, error)`, `Algorithm() string`, and `DID(ctx) (string, error)`) instead of `wallet.WalletProvider`. The `internal/p2p/handshake` package SHALL NOT import `internal/wallet`. The `Challenge` and `ChallengeResponse` structs SHALL include a `SignatureAlgorithm` field (omitempty for backward compatibility). The `Config` SHALL include `LegacySigner Signer` for v1 fallback. The `Handshaker` SHALL select signer based on peer's algorithm via `selectSigner(peerAlgo)` and dispatch signature verification based on the `SignatureAlgorithm` field, defaulting to `"secp256k1-keccak256"` when empty. The responder SHALL use `signer.DID(ctx)` to populate the ChallengeResponse DID field.

#### Scenario: Signer interface replaces wallet dependency
- **WHEN** `NewHandshaker` is called with a `Config` containing a `Signer`
- **THEN** the handshaker SHALL use only `SignMessage` and `PublicKey` from the signer
- **AND** `internal/p2p/handshake` SHALL NOT have an import path to `internal/wallet`

#### Scenario: Signer provides DID directly
- **WHEN** `HandleIncoming` constructs a ChallengeResponse
- **THEN** it SHALL call `signer.DID(ctx)` to get the DID string
- **AND** it SHALL NOT call `identity.DIDFromPublicKey` directly

#### Scenario: LegacySigner used for unknown peers
- **WHEN** `Initiate` is called to connect to an unknown peer
- **THEN** the handshaker SHALL use `LegacySigner` (secp256k1) for signing

#### Scenario: Responder matches initiator algorithm
- **WHEN** `HandleIncoming` receives a challenge with `SignatureAlgorithm = "ed25519"`
- **THEN** the responder SHALL use the primary `Signer` (Ed25519)

#### Scenario: Responder falls back for v1 initiator
- **WHEN** `HandleIncoming` receives a challenge with empty `SignatureAlgorithm`
- **THEN** the responder SHALL use `LegacySigner` (secp256k1)

#### Scenario: Successful handshake with ECDSA signature
- **WHEN** `Handshaker.Initiate` is called with `ZKEnabled=false` and the remote peer completes the challenge-response
- **THEN** `Initiate` SHALL return a valid `*Session` with `ZKVerified=false` and the remote DID populated

#### Scenario: Successful handshake with ZK proof
- **WHEN** `Handshaker.Initiate` is called with `ZKEnabled=true` and the remote peer returns a ZK proof
- **THEN** `Initiate` SHALL call the `ZKVerifierFunc`, and if valid, return a `*Session` with `ZKVerified=true`

#### Scenario: ZK proof verification failure rejects handshake
- **WHEN** the `ZKVerifierFunc` returns `false` for the received ZK proof
- **THEN** `Handshaker.Initiate` SHALL return an error containing "ZK proof invalid"

#### Scenario: Nonce mismatch rejects response
- **WHEN** the `ChallengeResponse` nonce differs from the nonce in the `Challenge`
- **THEN** `verifyResponse` SHALL return an error containing "nonce mismatch" using constant-time comparison (`hmac.Equal`)

#### Scenario: Valid ECDSA signature accepted
- **WHEN** a challenge response contains a 65-byte ECDSA signature that recovers to a public key matching `resp.PublicKey`
- **THEN** the verifier SHALL accept the response as authenticated

#### Scenario: Invalid signature rejected (public key mismatch)
- **WHEN** a challenge response contains a signature that recovers to a public key NOT matching `resp.PublicKey`
- **THEN** the verifier SHALL reject the response with "signature public key mismatch" error

#### Scenario: Wrong signature length rejected
- **WHEN** a challenge response contains a signature that is not exactly 65 bytes
- **THEN** the verifier SHALL reject the response with "invalid signature length" error

#### Scenario: Corrupted signature rejected
- **WHEN** a challenge response contains a 65-byte signature that cannot be recovered to a valid public key
- **THEN** the verifier SHALL reject the response with an error

#### Scenario: Response with neither proof nor signature rejected
- **WHEN** the `ChallengeResponse` has empty `ZKProof` and empty `Signature`
- **THEN** `verifyResponse` SHALL return an error containing "no proof or signature in response"

#### Scenario: Handshake timeout enforced
- **WHEN** the remote peer does not respond within `cfg.Timeout` duration
- **THEN** `Handshaker.Initiate` SHALL return a context deadline exceeded error

---

### Requirement: Human-in-the-Loop (HITL) Approval on Incoming Handshake

When a peer initiates an incoming handshake, the `Handshaker.HandleIncoming` method MUST invoke the `ApprovalFunc` before sending a response. If the user denies approval, the handshake SHALL be rejected with an error containing "handshake denied by user". Known peers with an active unexpired session MAY be auto-approved if `AutoApproveKnown=true`.

#### Scenario: New peer requires user approval
- **WHEN** `HandleIncoming` is called and no existing session exists for the sender's DID
- **THEN** `ApprovalFunc` SHALL be called with a `PendingHandshake` containing the peer ID, DID, remote address, and timestamp

#### Scenario: User denies incoming handshake
- **WHEN** the `ApprovalFunc` returns `(false, nil)`
- **THEN** `HandleIncoming` SHALL return an error containing "handshake denied by user" and SHALL NOT send a response

#### Scenario: Known peer with AutoApproveKnown skips approval
- **WHEN** `HandleIncoming` is called, `AutoApproveKnown=true`, and a valid session already exists for the sender's DID
- **THEN** `ApprovalFunc` SHALL NOT be called and the handshake SHALL proceed directly to response generation

#### Scenario: ApprovalFunc error propagates
- **WHEN** `ApprovalFunc` returns a non-nil error
- **THEN** `HandleIncoming` SHALL return a wrapped error and SHALL NOT proceed with the handshake

---

### Requirement: ZK Proof Fallback to Signature

When `ZKEnabled=true` but the `ZKProverFunc` returns an error, `HandleIncoming` SHALL fall back to ECDSA wallet signature. The fallback MUST be logged as a warning. The response SHALL contain the signature in the `Signature` field with `ZKProof` empty.

#### Scenario: ZK prover failure triggers signature fallback
- **WHEN** `ZKProverFunc` returns an error during `HandleIncoming`
- **THEN** the handler SHALL log a warning, call `wallet.SignMessage` with the challenge nonce, and set `resp.Signature`

#### Scenario: Signature fallback failure rejects handshake
- **WHEN** `ZKProverFunc` fails AND `wallet.SignMessage` also returns an error
- **THEN** `HandleIncoming` SHALL return a wrapped error containing "sign challenge"

---

### Requirement: Constant-time nonce comparison
The handshake verifier SHALL use `hmac.Equal()` for nonce comparison to prevent timing side-channel attacks.

#### Scenario: Nonce mismatch detected securely
- **WHEN** the response nonce does not match the challenge nonce
- **THEN** the verifier SHALL reject the response with "nonce mismatch" error using constant-time comparison

---

### Requirement: Bundle transport in handshake

The `Challenge` and `ChallengeResponse` structs SHALL include an optional `Bundle *IdentityBundle` field (omitempty). V2 signers SHALL include their IdentityBundle in handshake messages. Received bundles SHALL be cached in the BundleResolver.

#### Scenario: v2 responder includes bundle
- **WHEN** a v2 responder sends a ChallengeResponse
- **THEN** `resp.Bundle` SHALL contain the responder's IdentityBundle

#### Scenario: v1 peer ignores bundle field
- **WHEN** a v1 peer receives a message with a Bundle field
- **THEN** the unknown field SHALL be ignored (JSON flexibility)

#### Scenario: Received bundle cached
- **WHEN** a handshake message with a Bundle is received
- **THEN** the bundle SHALL be stored in the BundleResolver cache

---

### Requirement: Handshake authentication

#### Scenario: Bundle cached only after authentication
- **WHEN** a handshake challenge or response is received
- **THEN** the bundle SHALL be cached only after signature verification succeeds
- **AND** alias registration SHALL occur only after approval succeeds

#### Scenario: v2 DID requires bundle with matching signing key
- **WHEN** a handshake peer claims a `did:lango:v2:` DID
- **THEN** the handshake SHALL require a non-nil bundle
- **AND** `ComputeDIDv2(bundle) == SenderDID` SHALL be verified
- **AND** `bytes.Equal(PublicKey, Bundle.SigningKey.PublicKey)` SHALL be verified
- **AND** failure of any check SHALL reject the handshake

#### Scenario: v1 DID matches public key
- **WHEN** a handshake peer claims a v1 `did:lango:` DID and provides a public key
- **THEN** `DIDFromPublicKey(PublicKey).ID == SenderDID` SHALL be verified
- **AND** mismatch SHALL reject the handshake

#### Scenario: Auto-approve uses existing alias only
- **WHEN** `AutoApproveKnownPeers` is enabled
- **THEN** session lookup SHALL use `DIDAlias.CanonicalDID` only if an alias was previously registered
- **AND** `bundle.LegacyDID` SHALL NOT be used for session lookup (unverifiable)

---

### Requirement: Injectable response verifier

The `Handshaker` SHALL support an injectable `SignatureVerifyFunc` for signature verification. When `Config.Verifiers` is nil, the default verifier map SHALL contain only `"secp256k1-keccak256"` → `VerifySecp256k1Signature`. The default verifier SHALL be extracted into a named exported function.

#### Scenario: Default verifier preserves existing behavior
- **WHEN** `NewHandshaker` is called with `Config.Verifiers = nil`
- **THEN** the handshaker SHALL use `VerifySecp256k1Signature` as the default response verifier

#### Scenario: Custom verifier injected
- **WHEN** `NewHandshaker` is called with a non-nil `Config.Verifiers` map
- **THEN** the handshaker SHALL use the provided map for response signature verification

#### Scenario: Ed25519 in default verifier map
- **WHEN** `NewHandshaker` is called with nil `Verifiers`
- **THEN** the default map SHALL include both `"secp256k1-keccak256"` and `"ed25519"` verifiers

#### Scenario: Signer declares algorithm
- **WHEN** `Initiate` constructs a Challenge
- **THEN** `challenge.SignatureAlgorithm` SHALL be set to `h.signer.Algorithm()`

#### Scenario: Challenge SenderDID from signer identity
- **WHEN** `Initiate()` constructs a Challenge
- **THEN** `challenge.SenderDID` SHALL be set to `initSigner.DID(ctx)` (the selected signer's DID), not the raw `localDID` parameter

#### Scenario: Responder declares algorithm
- **WHEN** `HandleIncoming` constructs a ChallengeResponse
- **THEN** `resp.SignatureAlgorithm` SHALL be set to `h.signer.Algorithm()`

#### Scenario: Backward compatible empty algorithm
- **WHEN** a Challenge or ChallengeResponse has an empty `SignatureAlgorithm`
- **THEN** the verifier SHALL default to `"secp256k1-keccak256"`

#### Scenario: Unsupported algorithm rejected
- **WHEN** the `SignatureAlgorithm` is not registered in the handshaker's verifier map
- **THEN** verification SHALL return an error containing "unsupported"

---

### Requirement: SignatureVerifyFunc type

The `SignatureVerifyFunc` type SHALL be used for both challenge and response signature verification, replacing the previous `ResponseVerifyFunc` name.

---

### Requirement: Challenge canonical payload

The `challengeCanonicalPayload` function SHALL return raw canonical bytes (nonce || bigEndian(timestamp) || senderDID) WITHOUT Keccak256 hashing. The signing and verification sides SHALL each hash the canonical payload once via their respective algorithm implementations, ensuring consistent hash depth.

#### Scenario: Challenge signature roundtrip succeeds
- **WHEN** a challenge is signed via `signer.SignMessage(challengeCanonicalPayload(...))` and verified via `verifyChallengeSignature`
- **THEN** verification SHALL succeed (single hash on each side)

---

### Requirement: Signature verification
The handshake verifier SHALL perform full ECDSA secp256k1 signature verification by recovering the public key from the signature using `ethcrypto.SigToPub()` and comparing it with the claimed public key via `ethcrypto.CompressPubkey()`, instead of accepting any non-empty signature.

#### Scenario: Valid signature accepted
- **WHEN** a challenge response contains a 65-byte ECDSA signature that recovers to a public key matching `resp.PublicKey`
- **THEN** the verifier SHALL accept the response as authenticated

#### Scenario: Invalid signature rejected
- **WHEN** a challenge response contains a signature that recovers to a public key NOT matching `resp.PublicKey`
- **THEN** the verifier SHALL reject the response with "signature public key mismatch" error

#### Scenario: Wrong signature length rejected
- **WHEN** a challenge response contains a signature that is not exactly 65 bytes
- **THEN** the verifier SHALL reject the response with "invalid signature length" error

#### Scenario: Corrupted signature rejected
- **WHEN** a challenge response contains a 65-byte signature that cannot be recovered to a valid public key
- **THEN** the verifier SHALL reject the response with an error

#### Scenario: No proof or signature rejected
- **WHEN** a challenge response contains neither a ZK proof nor a signature
- **THEN** the verifier SHALL reject the response with "no proof or signature in response" error

---

### Requirement: Session Store with TTL Eviction

The `SessionStore` SHALL store authenticated peer sessions keyed by peer DID. Session tokens SHALL be generated as HMAC-SHA256 over random bytes and the peer DID using a 32-byte randomly generated HMAC key created at store initialization. Sessions SHALL have a configurable TTL. Expired sessions SHALL be evicted lazily on access and proactively via `Cleanup()`.

#### Scenario: Session created with correct fields
- **WHEN** `SessionStore.Create("did:lango:abc", true)` is called
- **THEN** a `Session` SHALL be stored with `PeerDID="did:lango:abc"`, `ZKVerified=true`, a non-empty `Token`, and `ExpiresAt = now + TTL`

#### Scenario: Valid session token validates successfully
- **WHEN** `SessionStore.Validate(peerDID, token)` is called with the correct peerDID and token from an unexpired session
- **THEN** `Validate` SHALL return `true`

#### Scenario: Expired session returns false on validation
- **WHEN** `SessionStore.Validate` is called and the session's `ExpiresAt` is in the past
- **THEN** `Validate` SHALL return `false` and SHALL remove the session from the store

#### Scenario: Session cleanup removes all expired entries
- **WHEN** `SessionStore.Cleanup()` is called
- **THEN** all sessions where `ExpiresAt` is before `time.Now()` SHALL be deleted and the count of removed sessions SHALL be returned

---

### Requirement: Hybrid PQ KEM key exchange in handshake v1.2

The `Handshaker` SHALL support an optional hybrid post-quantum KEM (X25519-MLKEM768) key exchange during handshake, enabled via `Config.EnablePQKEM`. When enabled, the initiator SHALL generate an ephemeral KEM keypair per handshake and include the public key in the Challenge. The responder SHALL encapsulate a shared secret using the initiator's KEM public key and include the ciphertext in the ChallengeResponse. Both sides SHALL derive a 32-byte session encryption key via `HKDF-SHA256(hybrid_shared_secret, nil, "lango-p2p-session-v1:" + initiatorSignerDID + ":" + responderSignerDID)`. The session key SHALL be stored in `Session.EncryptionKey` (never serialized) with `Session.KEMUsed = true`.

#### Scenario: Full KEM exchange between v1.2 peers
- **WHEN** both initiator and responder have `EnablePQKEM = true`
- **THEN** the initiator SHALL include `KEMPublicKey` and `KEMAlgorithm` in the Challenge
- **AND** the responder SHALL encapsulate and include `KEMCiphertext` in the ChallengeResponse
- **AND** both sides SHALL derive identical 32-byte session encryption keys
- **AND** `Session.KEMUsed` SHALL be `true`

#### Scenario: v1.2 initiator with v1.1 responder (graceful degradation)
- **WHEN** the initiator has `EnablePQKEM = true` but the responder does not support KEM
- **THEN** the ChallengeResponse SHALL have empty `KEMCiphertext`
- **AND** `Session.KEMUsed` SHALL be `false`
- **AND** the handshake SHALL succeed with signature-only authentication

#### Scenario: v1.1 initiator with v1.2 responder
- **WHEN** the initiator does not include KEM fields in the Challenge
- **THEN** the responder SHALL skip KEM encapsulation
- **AND** the handshake SHALL proceed as a normal v1.1 handshake

#### Scenario: KEM keypair generation failure
- **WHEN** `GenerateEphemeralKEM()` returns an error
- **THEN** the initiator SHALL log a warning and proceed without KEM fields (v1.1 fallback)

#### Scenario: KEM encapsulation failure
- **WHEN** `KEMEncapsulate()` returns an error on the responder side
- **THEN** the responder SHALL log a warning, omit `KEMCiphertext`, and proceed without KEM

---

### Requirement: KEM transcript binding in v1.2 signatures

The v1.2 challenge canonical payload SHALL include `kemAlgorithm` and `kemPublicKey` appended after `senderDID`. The v1.2 response canonical payload SHALL include `kemCiphertext` appended after the nonce. When KEM fields are empty (v1.1 messages), the canonical payloads SHALL be identical to v1.1 for backward compatibility.

#### Scenario: Challenge signature covers KEM public key
- **WHEN** a v1.2 challenge is signed
- **THEN** the canonical payload SHALL be `nonce || bigEndian(timestamp) || utf8(senderDID) || utf8(kemAlgorithm) || kemPublicKey`

#### Scenario: Response signature covers KEM ciphertext
- **WHEN** a v1.2 response is signed
- **THEN** the responder SHALL sign `nonce || kemCiphertext` instead of raw nonce

#### Scenario: Tampered KEM public key rejected
- **WHEN** an attacker modifies `KEMPublicKey` in a v1.2 challenge
- **THEN** `verifyChallengeSignature()` SHALL reject the challenge due to signature mismatch

#### Scenario: Tampered KEM ciphertext rejected
- **WHEN** an attacker modifies `KEMCiphertext` in a v1.2 response
- **THEN** `verifyResponse()` SHALL reject the response due to signature mismatch

#### Scenario: v1.1 canonical payload unchanged
- **WHEN** KEM fields are empty (v1.1 message)
- **THEN** the canonical payload SHALL be identical to the current v1.1 format

---

### Requirement: Session encryption key zeroing

`Session.EncryptionKey` SHALL be zeroed via `security.ZeroBytes()` at all exit points: `Create()` overwrite (when same peerDID session already exists), `Remove()`, `Invalidate()`, `InvalidateAll()`, `InvalidateByCondition()`, and `Cleanup()`.

#### Scenario: Key zeroed on session removal
- **WHEN** `SessionStore.Remove(peerDID)` is called and the session has an EncryptionKey
- **THEN** the EncryptionKey SHALL be zeroed before the session is deleted from the map

#### Scenario: Key zeroed on session overwrite
- **WHEN** `SessionStore.Create(peerDID, ...)` is called and a session for that peerDID already exists
- **THEN** the existing session's EncryptionKey SHALL be zeroed before being replaced

---

### Requirement: Initiator protocol selection

The handshake package SHALL provide a `PreferredProtocols(kemEnabled bool)` function returning the ordered list of protocol IDs to try. When `kemEnabled = true`, the order SHALL be `[v1.2, v1.1, v1.0]`. When `kemEnabled = false`, the order SHALL be `[v1.1, v1.0]`. The initiator call site SHALL use this function with libp2p's `NewStream` for automatic fallback.

#### Scenario: KEM enabled initiator
- **WHEN** `PreferredProtocols(true)` is called
- **THEN** it SHALL return `["/lango/handshake/1.2.0", "/lango/handshake/1.1.0", "/lango/handshake/1.0.0"]`

#### Scenario: KEM disabled initiator
- **WHEN** `PreferredProtocols(false)` is called
- **THEN** it SHALL return `["/lango/handshake/1.1.0", "/lango/handshake/1.0.0"]`
