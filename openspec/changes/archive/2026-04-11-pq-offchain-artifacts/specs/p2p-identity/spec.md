## MODIFIED Requirements

### Requirement: IdentityBundle type

#### Scenario: Bundle with ML-DSA-65 PQ signing key
- **WHEN** a new IdentityBundle is created with a PQ signing key available
- **THEN** `PQSigningKey.Algorithm` SHALL be "ml-dsa-65" and `PQGeneration` SHALL reflect the ML-DSA key derivation generation
- **AND** `BundleProofs.MLDSA65` SHALL contain an ML-DSA-65 signature over the canonical bundle bytes
- **AND** `PQSigningKey` and `PQGeneration` SHALL NOT be included in `CanonicalBundleBytes` (DID v2 hash unchanged)

#### Scenario: Bundle without PQ key (backward compat)
- **WHEN** a legacy IdentityBundle is deserialized without PQ fields
- **THEN** `PQSigningKey` SHALL be nil and `PQGeneration` SHALL be 0
- **AND** the bundle SHALL be fully functional for v1/v2 DID operations

### Requirement: Identity key derivation from Master Key

#### Scenario: PQ signing key derived from MK
- **WHEN** `DerivePQSigningKey(mk, generation)` is called
- **THEN** it SHALL derive a 32-byte seed via `HKDF-SHA256(mk, nil, "lango-pq-signing-mldsa65[:generation]")`
- **AND** produce a deterministic ML-DSA-65 keypair via `mldsa65.NewKeyFromSeed(seed)`
- **AND** the derived key SHALL be independent from the Ed25519 identity key (different HKDF domain)

#### Scenario: Same MK produces same PQ key
- **WHEN** `DerivePQSigningKey(mk, 0)` is called twice with the same MK
- **THEN** both calls SHALL return identical ML-DSA-65 private keys
