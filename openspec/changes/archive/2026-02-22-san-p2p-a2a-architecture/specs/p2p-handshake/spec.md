## ADDED Requirements

### Requirement: Challenge-Response Mutual Authentication

The `Handshaker` SHALL implement a three-message challenge-response protocol over libp2p streams using protocol ID `/lango/handshake/1.0.0`. The initiator SHALL send a `Challenge` containing a 32-byte cryptographically random nonce, a Unix timestamp, and the sender's DID. The responder SHALL reply with a `ChallengeResponse` containing the echoed nonce, the responder's DID, the responder's compressed public key, and either a ZK proof or an ECDSA signature. The initiator SHALL send a `SessionAck` containing the session token and expiry on successful verification.

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
- **THEN** `verifyResponse` SHALL return an error containing "nonce mismatch"

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
