## Tasks

### Wave 1 — ML-DSA-65 primitives

- [x] Add `AlgorithmMLDSA65` constant to `internal/security/signature_scheme.go`
- [x] Create `internal/security/scheme_mldsa65.go` — `VerifyMLDSA65()`, `SignMLDSA65()`, `MLDSA65Scheme`
- [x] Create `internal/security/scheme_mldsa65_test.go` — sign/verify roundtrip, invalid key/sig rejection
- [x] Create `internal/security/derive_pq.go` — `DerivePQSigningKey(mk, generation)`
- [x] Create `internal/security/derive_pq_test.go` — determinism, generation rotation, domain separation from Ed25519

### Wave 2 — Identity bundle PQ extension + bootstrap

- [x] Add `PQGeneration`, `PQSigningKey` to IdentityBundle, `MLDSA65` to BundleProofs (CanonicalBundleBytes unchanged)
- [x] Add PQ key seed to BundleProviderConfig, derive ML-DSA key, create PQ proof
- [x] Implement `PQBundleSigner` interface on BundleProvider (SignPQ + PQAlgorithm)
- [x] Add `PQSigningKeySeed` to bootstrap State + Result
- [x] Add `phaseDerivePQKey()` bootstrap phase
- [x] Test: DID v2 hash unchanged with PQ key, PQ proof roundtrip

### Wave 3 — Provenance dual signatures

- [x] Add `PQSignerPublicKey`, `PQSignatureAlgorithm`, `PQSignature` to ProvenanceBundle
- [x] Add `PQBundleSigner` interface to provenance package
- [x] Update `Export()` for dual-sign (classical + PQ)
- [x] Update `canonicalBundlePayload()` to zero both Signature and PQSignature (include PQSignerPublicKey)
- [x] Update `Verify()` for dual-verify
- [x] Add ML-DSA-65 verifier to `modules_provenance.go` and `cli/provenance/common.go`
- [x] Test: dual-signature roundtrip, classical-only backward compat

### Wave 4 — GossipCard signing

- [x] Add signature fields to GossipCard struct
- [x] Add `CanonicalCardPayload()` (includes Bundle, excludes Signature/PQSignature only)
- [x] Sign card on publish, verify on receive (accept unsigned for backward compat)
- [x] Wire card signer in `wiring_p2p.go`

### Wave 5 — Wiring + CLI

- [x] Pass `boot.PQSigningKeySeed` to BundleProviderConfig in `wiring_p2p.go`
- [x] Wire PQ signer for provenance export in `wiring_provenance.go`
- [x] Add PQ signing key status to `cli/security/status.go`

### Wave 6 — OpenSpec finalize

- [x] Verify change
- [x] Sync delta specs
- [x] Archive change
