## MODIFIED Requirements

### Requirement: Challenge-Response Mutual Authentication

The `Handshaker` SHALL accept a `Signer` interface (methods `SignMessage(ctx, message) ([]byte, error)` and `PublicKey(ctx) ([]byte, error)`) instead of `wallet.WalletProvider`. The `internal/p2p/handshake` package SHALL NOT import `internal/wallet`. The responder SHALL derive its DID using `identity.DIDFromPublicKey(pubkey)` instead of inline string construction.

#### Scenario: Signer interface replaces wallet dependency
- **WHEN** `NewHandshaker` is called with a `Config` containing a `Signer`
- **THEN** the handshaker SHALL use only `SignMessage` and `PublicKey` from the signer
- **AND** `internal/p2p/handshake` SHALL NOT have an import path to `internal/wallet`

#### Scenario: Responder DID derived via identity.DIDFromPublicKey
- **WHEN** `HandleIncoming` constructs the responder's DID for the `ChallengeResponse`
- **THEN** it SHALL call `identity.DIDFromPublicKey(pubkey)` and use `did.ID`
- **AND** it SHALL NOT construct the DID string inline

### Requirement: Injectable response verifier

The `Handshaker` SHALL support an injectable `ResponseVerifyFunc` for signature verification. When `Config.ResponseVerifier` is nil, the default secp256k1+keccak256 verification SHALL be used. The default verifier SHALL be extracted into a named exported function (`VerifySecp256k1Signature`).

#### Scenario: Default verifier preserves existing behavior
- **WHEN** `NewHandshaker` is called with `Config.ResponseVerifier = nil`
- **THEN** the handshaker SHALL use `VerifySecp256k1Signature` as the response verifier

#### Scenario: Custom verifier injected
- **WHEN** `NewHandshaker` is called with a non-nil `Config.ResponseVerifier`
- **THEN** the handshaker SHALL use the provided function for response signature verification
