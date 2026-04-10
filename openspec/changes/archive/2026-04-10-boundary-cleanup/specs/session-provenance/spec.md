## MODIFIED Requirements

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
