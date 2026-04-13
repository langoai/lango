## ADDED Requirements

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
