## MODIFIED Requirements

### Requirement: Signature verification uses injected verifiers

The wiring layer SHALL register Ed25519 as an additional verifier alongside secp256k1-keccak256. The Ed25519 verifier SHALL be constructed as a wiring closure combining `identity.ParseDIDPublicKey` and `security.VerifyEd25519`. Ed25519 provenance verification is a framework validation capability, not a production user feature in Phase 2.

#### Scenario: Ed25519 provenance bundle verifiable
- **WHEN** a bundle is exported with `SignatureAlgorithm = "ed25519"` and verified
- **THEN** the verifier map SHALL dispatch to the Ed25519 verifier closure
- **AND** verification SHALL succeed if the signature is valid

### Requirement: Algorithm constant canonical source

The algorithm constant `AlgorithmSecp256k1Keccak256` SHALL be defined in `internal/security` as the canonical source. The `internal/provenance` package SHALL re-export it for backward compatibility.
