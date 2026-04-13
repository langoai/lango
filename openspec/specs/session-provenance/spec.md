## Purpose

Capability spec for session-provenance. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Config fingerprint in provenance checkpoint
The provenance subsystem SHALL compute a SHA-256 fingerprint of session-relevant configuration state (explicit keys, auto-enabled flags, hooks config) and store it as `config_fingerprint` metadata in session provenance checkpoints.

#### Scenario: Config fingerprint recorded at session start
- **WHEN** the provenance module initializes for a session
- **THEN** a provenance checkpoint SHALL include a `config_fingerprint` metadata field containing a hex-encoded SHA-256 digest of the serialized config state

### Requirement: Hook registry snapshot in provenance checkpoint
The provenance subsystem SHALL capture a JSON snapshot of the current hook registry (pre-hooks and post-hooks with name and priority) and store it as `hook_registry` metadata in session provenance checkpoints.

#### Scenario: Hook snapshot recorded in checkpoint metadata
- **WHEN** a session provenance checkpoint is created
- **THEN** the checkpoint metadata SHALL include a `hook_registry` field containing a JSON array of hook entries, each with `name` and `priority` fields

### Requirement: Bundle signing uses BundleSigner interface

The `BundleService.Export` method SHALL accept a `BundleSigner` interface (methods `Sign(ctx, payload) ([]byte, error)` and `Algorithm() string`) instead of `BundleSignFunc`. The signature algorithm in the exported bundle SHALL be set from `signer.Algorithm()`, not from a hardcoded constant.

#### Scenario: BundleSigner provides algorithm
- **WHEN** `Export` is called with a `BundleSigner` whose `Algorithm()` returns `"secp256k1-keccak256"`
- **THEN** the bundle's `SignatureAlgorithm` field SHALL be `"secp256k1-keccak256"`
- **AND** the signature SHALL be produced by calling `signer.Sign(ctx, payload)`

### Requirement: Signature verification uses injected verifiers

The `BundleService` SHALL receive a `map[string]SignatureVerifyFunc` at construction time. `Verify` SHALL look up the bundle's `SignatureAlgorithm` in this map and call the corresponding verifier. The `internal/provenance` package SHALL NOT import `internal/p2p/identity`. Verification implementation is owned by the `app/cli` integration layer.

#### Scenario: Verifier dispatched by algorithm
- **WHEN** `Verify` is called on a bundle with `SignatureAlgorithm = "secp256k1-keccak256"`
- **THEN** the verifier registered for that algorithm key SHALL be called

#### Scenario: Unknown algorithm rejected
- **WHEN** `Verify` is called on a bundle with an unregistered `SignatureAlgorithm`
- **THEN** `Verify` SHALL return an error containing "unsupported signature algorithm"

#### Scenario: No default verifier in provenance package
- **WHEN** `NewBundleService` is called with an empty verifiers map
- **THEN** all `Verify` calls SHALL return "unsupported signature algorithm" errors
- **AND** the provenance package SHALL NOT contain any hardcoded verifier implementation

#### Scenario: Ed25519 provenance bundle verifiable
- **WHEN** a bundle is exported with `SignatureAlgorithm = "ed25519"` and verified
- **THEN** the verifier map SHALL dispatch to the Ed25519 verifier closure
- **AND** verification SHALL succeed if the signature is valid

### Requirement: Algorithm constant canonical source

The algorithm constant `AlgorithmSecp256k1Keccak256` SHALL be defined in `internal/security` as the canonical source. The `internal/provenance` package SHALL re-export it for backward compatibility.

---

### Requirement: PQ dual signatures on provenance bundles

The `BundleService.Export()` SHALL generate both a classical signature and an optional ML-DSA-65 PQ signature on provenance bundles. The PQ signature is generated when the signer implements `PQBundleSigner`. The `PQSignerPublicKey` SHALL be embedded in the bundle for self-contained, rotation-safe verification.

#### Scenario: Dual-signed provenance bundle
- **WHEN** `Export()` is called with a signer that implements `PQBundleSigner`
- **THEN** the bundle SHALL contain both `Signature` (classical) and `PQSignature` (ML-DSA-65)
- **AND** `PQSignerPublicKey` SHALL be embedded in the bundle
- **AND** `PQSignatureAlgorithm` SHALL be "ml-dsa-65"

#### Scenario: Classical-only provenance bundle (backward compat)
- **WHEN** `Export()` is called with a signer that does NOT implement `PQBundleSigner`
- **THEN** the bundle SHALL contain only `Signature` (classical)
- **AND** `PQSignature`, `PQSignerPublicKey`, `PQSignatureAlgorithm` SHALL be empty

#### Scenario: Verify dual-signed bundle
- **WHEN** `Verify()` is called on a bundle with both classical and PQ signatures
- **THEN** classical signature SHALL be verified against the DID (required)
- **AND** PQ signature SHALL be verified against the embedded `PQSignerPublicKey`

#### Scenario: Verify classical-only bundle
- **WHEN** `Verify()` is called on a bundle without PQ signature
- **THEN** classical signature SHALL be verified (required)
- **AND** PQ verification SHALL be skipped

#### Scenario: PQ signature covers PQSignerPublicKey
- **WHEN** the canonical bundle payload is constructed
- **THEN** `Signature` and `PQSignature` SHALL be excluded
- **AND** `PQSignerPublicKey` and `PQSignatureAlgorithm` SHALL be included
- **AND** classical signature authenticates the embedded PQ public key
