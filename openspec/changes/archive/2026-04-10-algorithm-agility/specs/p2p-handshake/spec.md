## MODIFIED Requirements

### Requirement: Challenge-Response Mutual Authentication

The `Signer` interface SHALL include an `Algorithm() string` method in addition to `SignMessage` and `PublicKey`. The `Challenge` and `ChallengeResponse` structs SHALL include a `SignatureAlgorithm` field (omitempty for backward compatibility). The `Handshaker` SHALL dispatch signature verification based on the `SignatureAlgorithm` field, defaulting to `"secp256k1-keccak256"` when empty.

#### Scenario: Signer declares algorithm
- **WHEN** `Initiate` constructs a Challenge
- **THEN** `challenge.SignatureAlgorithm` SHALL be set to `h.signer.Algorithm()`

#### Scenario: Responder declares algorithm
- **WHEN** `HandleIncoming` constructs a ChallengeResponse
- **THEN** `resp.SignatureAlgorithm` SHALL be set to `h.signer.Algorithm()`

#### Scenario: Backward compatible empty algorithm
- **WHEN** a Challenge or ChallengeResponse has an empty `SignatureAlgorithm`
- **THEN** the verifier SHALL default to `"secp256k1-keccak256"`

#### Scenario: Unsupported algorithm rejected
- **WHEN** the `SignatureAlgorithm` is not registered in the handshaker's verifier map
- **THEN** verification SHALL return an error containing "unsupported"

#### Scenario: Verifier map replaces single verifier
- **WHEN** `NewHandshaker` is called with a nil `Verifiers` map
- **THEN** the default map SHALL contain only `"secp256k1-keccak256"` → `VerifySecp256k1Signature`

### Requirement: SignatureVerifyFunc type rename

The `ResponseVerifyFunc` type SHALL be renamed to `SignatureVerifyFunc` to reflect its use in both challenge and response verification.

### Requirement: Challenge canonical payload (bug fix)

The `challengeCanonicalPayload` function SHALL return raw canonical bytes (nonce || bigEndian(timestamp) || senderDID) WITHOUT Keccak256 hashing. The signing and verification sides SHALL each hash the canonical payload once via their respective algorithm implementations, ensuring consistent hash depth.

#### Scenario: Challenge signature roundtrip succeeds
- **WHEN** a challenge is signed via `signer.SignMessage(challengeCanonicalPayload(...))` and verified via `verifyChallengeSignature`
- **THEN** verification SHALL succeed (single hash on each side)
