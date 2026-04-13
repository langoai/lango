## MODIFIED Requirements

### Requirement: Challenge-Response Mutual Authentication

The `Signer` interface SHALL include a `DID(ctx) (string, error)` method. `HandleIncoming` SHALL use `signer.DID(ctx)` instead of `identity.DIDFromPublicKey(pubkey)` to populate the ChallengeResponse DID field. The `Config` SHALL include `LegacySigner Signer` for v1 fallback. The `Handshaker` SHALL select signer based on peer's algorithm via `selectSigner(peerAlgo)`.

#### Scenario: Signer provides DID directly
- **WHEN** `HandleIncoming` constructs a ChallengeResponse
- **THEN** it SHALL call `signer.DID(ctx)` to get the DID string
- **AND** it SHALL NOT call `identity.DIDFromPublicKey` directly

#### Scenario: LegacySigner used for unknown peers
- **WHEN** `Initiate` is called to connect to an unknown peer
- **THEN** the handshaker SHALL use `LegacySigner` (secp256k1) for signing

#### Scenario: Primary signer used for known v2 peers
- **WHEN** `Initiate` is called to connect to a known v2 peer (ed25519 algorithm)
- **THEN** the handshaker SHALL use the primary `Signer` (Ed25519)

#### Scenario: Responder matches initiator algorithm
- **WHEN** `HandleIncoming` receives a challenge with `SignatureAlgorithm = "ed25519"`
- **THEN** the responder SHALL use the primary `Signer` (Ed25519)

#### Scenario: Responder falls back for v1 initiator
- **WHEN** `HandleIncoming` receives a challenge with empty `SignatureAlgorithm`
- **THEN** the responder SHALL use `LegacySigner` (secp256k1)

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

### Requirement: Injectable response verifier

Ed25519 SHALL be included in the default verifier map alongside secp256k1-keccak256.

#### Scenario: Ed25519 in default verifier map
- **WHEN** `NewHandshaker` is called with nil `Verifiers`
- **THEN** the default map SHALL include both `"secp256k1-keccak256"` and `"ed25519"` verifiers
