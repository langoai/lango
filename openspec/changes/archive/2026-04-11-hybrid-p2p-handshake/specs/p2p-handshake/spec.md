## ADDED Requirements

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

### Requirement: Session encryption key zeroing

`Session.EncryptionKey` SHALL be zeroed via `security.ZeroBytes()` at all exit points: `Create()` overwrite (when same peerDID session already exists), `Remove()`, `Invalidate()`, `InvalidateAll()`, `InvalidateByCondition()`, and `Cleanup()`.

#### Scenario: Key zeroed on session removal
- **WHEN** `SessionStore.Remove(peerDID)` is called and the session has an EncryptionKey
- **THEN** the EncryptionKey SHALL be zeroed before the session is deleted from the map

#### Scenario: Key zeroed on session overwrite
- **WHEN** `SessionStore.Create(peerDID, ...)` is called and a session for that peerDID already exists
- **THEN** the existing session's EncryptionKey SHALL be zeroed before being replaced

### Requirement: Initiator protocol selection

The handshake package SHALL provide a `PreferredProtocols(kemEnabled bool)` function returning the ordered list of protocol IDs to try. When `kemEnabled = true`, the order SHALL be `[v1.2, v1.1, v1.0]`. When `kemEnabled = false`, the order SHALL be `[v1.1, v1.0]`. The initiator call site SHALL use this function with libp2p's `NewStream` for automatic fallback.

#### Scenario: KEM enabled initiator
- **WHEN** `PreferredProtocols(true)` is called
- **THEN** it SHALL return `["/lango/handshake/1.2.0", "/lango/handshake/1.1.0", "/lango/handshake/1.0.0"]`

#### Scenario: KEM disabled initiator
- **WHEN** `PreferredProtocols(false)` is called
- **THEN** it SHALL return `["/lango/handshake/1.1.0", "/lango/handshake/1.0.0"]`

## MODIFIED Requirements

### Requirement: Challenge-Response Mutual Authentication

#### Scenario: Challenge SenderDID from signer identity
- **WHEN** `Initiate()` constructs a Challenge
- **THEN** `challenge.SenderDID` SHALL be set to `initSigner.DID(ctx)` (the selected signer's DID), not the raw `localDID` parameter
