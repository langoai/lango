## Purpose

Capability spec for p2p-identity. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: DID Derivation from Public Key

The `WalletDIDProvider` SHALL accept a `KeyProvider` interface (single method `PublicKey(ctx) ([]byte, error)`) instead of `wallet.WalletProvider`. The `internal/p2p/identity` package SHALL NOT import `internal/wallet`. The `wallet.WalletProvider` implicitly satisfies `KeyProvider` via Go structural typing. The DID format SHALL be `did:lango:<hex-encoded-compressed-pubkey>`. The derived DID SHALL be cached after the first derivation.

#### Scenario: KeyProvider interface replaces wallet dependency
- **WHEN** `NewProvider` is called with any type satisfying `KeyProvider`
- **THEN** the provider SHALL use only `PublicKey(ctx)` to derive the DID
- **AND** `internal/p2p/identity` SHALL NOT have an import path to `internal/wallet`

#### Scenario: DID derived on first call
- **WHEN** `WalletDIDProvider.DID(ctx)` is called for the first time
- **THEN** the provider SHALL call `keys.PublicKey(ctx)`, construct a DID with prefix `did:lango:`, encode the public key as lowercase hex, and cache the result

#### Scenario: DID returned from cache on subsequent calls
- **WHEN** `WalletDIDProvider.DID(ctx)` is called after a successful first call
- **THEN** the provider SHALL return the cached DID without calling `wallet.PublicKey` again

#### Scenario: Wallet public key error propagates
- **WHEN** `wallet.PublicKey(ctx)` returns an error
- **THEN** `WalletDIDProvider.DID(ctx)` SHALL return a nil DID and a wrapped error; the cache SHALL NOT be populated

---

### Requirement: Peer ID Derivation from Public Key

The system SHALL derive a libp2p `peer.ID` from a public key by detecting the key type from its size: 33-byte compressed secp256k1 via `crypto.UnmarshalSecp256k1PublicKey`, 32-byte Ed25519 via `crypto.UnmarshalEd25519PublicKey`. The derived `peer.ID` SHALL be embedded in the `DID` struct. This mapping SHALL be deterministic: the same public key always produces the same peer ID.

#### Scenario: Valid compressed public key produces peer ID
- **WHEN** `DIDFromPublicKey` is called with a valid 33-byte compressed secp256k1 public key
- **THEN** a `DID` struct SHALL be returned with a non-empty `PeerID` field derived from the key

#### Scenario: Empty public key rejected
- **WHEN** `DIDFromPublicKey` is called with an empty byte slice
- **THEN** the function SHALL return an error containing "empty public key"

#### Scenario: Invalid public key bytes rejected
- **WHEN** `DIDFromPublicKey` is called with malformed bytes that are not a valid secp256k1 point
- **THEN** the function SHALL return an error from `crypto.UnmarshalSecp256k1PublicKey`

#### Scenario: Ed25519 public key produces peer ID
- **WHEN** `peerIDFromPublicKey` is called with a 32-byte Ed25519 public key
- **THEN** it SHALL derive a peer ID via `crypto.UnmarshalEd25519PublicKey`

---

### Requirement: DID Verification Against Peer ID

The `WalletDIDProvider.VerifyDID` method SHALL re-derive the `peer.ID` from the public key embedded in a `DID` struct and compare it to the claimed `peer.ID`. If they do not match, the method MUST return an error describing the mismatch. A nil DID MUST return an error.

#### Scenario: Valid DID matches peer ID
- **WHEN** `VerifyDID` is called with a DID whose public key was used to derive the provided peer ID
- **THEN** `VerifyDID` SHALL return nil (no error)

#### Scenario: DID public key does not match claimed peer ID
- **WHEN** `VerifyDID` is called with a DID whose public key produces a different peer ID than the one provided
- **THEN** `VerifyDID` SHALL return an error containing "peer ID mismatch"

#### Scenario: Nil DID rejected
- **WHEN** `VerifyDID` is called with a nil `DID` pointer
- **THEN** `VerifyDID` SHALL return an error containing "nil DID"

---

### Requirement: ParseDIDPublicKey helper

The system SHALL provide a `ParseDIDPublicKey` function that extracts raw public key bytes from a `did:lango:<hex>` string without deriving a peer ID. This is a read-only helper for signature verification where the caller knows the key type independently.

#### Scenario: ParseDIDPublicKey extracts bytes
- **WHEN** `ParseDIDPublicKey("did:lango:<valid-hex>")` is called
- **THEN** it SHALL return the hex-decoded byte slice without calling `peerIDFromPublicKey`

#### Scenario: ParseDIDPublicKey rejects invalid prefix
- **WHEN** `ParseDIDPublicKey` is called with a non-`did:lango:` prefix
- **THEN** it SHALL return an error

#### Scenario: ParseDIDPublicKey rejects empty key
- **WHEN** `ParseDIDPublicKey("did:lango:")` is called
- **THEN** it SHALL return an error containing "empty public key"

---

### Requirement: DID Parsing from String

`ParseDID` SHALL dispatch by prefix: `did:lango:v2:` to v2 parser, `did:lango:` to v1 parser. V1 parser validates the prefix, decodes the hex-encoded public key, and derives the peer ID. V2 parser returns DID{Version:2, ID:didStr} with empty PublicKey and PeerID (requires BundleResolver for full resolution). Any malformed input SHALL result in an error.

#### Scenario: Valid DID string parsed
- **WHEN** `ParseDID("did:lango:<valid-hex-pubkey>")` is called
- **THEN** the function SHALL return a `DID` struct with the correct `ID`, `PublicKey`, and `PeerID` fields

#### Scenario: Missing prefix rejected
- **WHEN** `ParseDID` is called with a string that does not start with `did:lango:`
- **THEN** the function SHALL return an error containing "invalid DID scheme"

#### Scenario: Empty key portion rejected
- **WHEN** `ParseDID("did:lango:")` is called with an empty hex key
- **THEN** the function SHALL return an error containing "empty public key in DID"

#### Scenario: Non-hex key portion rejected
- **WHEN** `ParseDID("did:lango:gg00ff")` is called with invalid hex characters
- **THEN** the function SHALL return an error from hex decoding

#### Scenario: v2 DID parsed with hollow fields
- **WHEN** `ParseDID("did:lango:v2:<hash>")` is called
- **THEN** the returned DID SHALL have Version=2, PublicKey=nil, PeerID=""

---

### Requirement: Identity command output
The `lango p2p identity` command SHALL display `keyStorage` information (either "secrets-store" or "file") instead of the raw `keyDir` filesystem path.

#### Scenario: Identity with encrypted storage
- **WHEN** the user runs `lango p2p identity` and SecretsStore is available
- **THEN** the output SHALL show `Key Storage: secrets-store` instead of a directory path

#### Scenario: Identity with file storage
- **WHEN** the user runs `lango p2p identity` and SecretsStore is not available
- **THEN** the output SHALL show `Key Storage: file`

#### Scenario: JSON output reflects key storage
- **WHEN** the user runs `lango p2p identity --json`
- **THEN** the JSON SHALL contain `"keyStorage": "secrets-store"` or `"keyStorage": "file"` instead of `"keyDir"`

---

### Requirement: Signature verification uses bytes.Equal

The `VerifyMessageSignature` function SHALL compare recovered public key bytes with DID public key bytes using `bytes.Equal`, not `string()` conversion.

#### Scenario: Byte comparison for signature verification
- **WHEN** `VerifyMessageSignature` compares the recovered public key with the DID's public key
- **THEN** it SHALL use `bytes.Equal(recovered, did.PublicKey)` for the comparison

---

### Requirement: DID v2 format

The system SHALL support a content-addressed DID format `did:lango:v2:<40-hex-chars>` where the identifier is SHA-256(canonical bundle bytes)[:20] hex-encoded. Canonical bundle bytes SHALL include Version, SigningKey, SettlementKey, and LegacyDID fields only (CreatedAt and Proofs excluded for determinism).

#### Scenario: DID v2 computed from bundle
- **WHEN** `ComputeDIDv2(bundle)` is called with an IdentityBundle
- **THEN** the result SHALL be `"did:lango:v2:" + hex(SHA256(canonical)[:20])`
- **AND** the same bundle always produces the same DID v2

---

### Requirement: IdentityBundle type

The system SHALL provide an `IdentityBundle` struct containing Version (int), SigningKey (PublicKeyEntry with Algorithm + PublicKey), SettlementKey (PublicKeyEntry), LegacyDID (string), Proofs (BundleProofs with Legacy + Ed25519 signatures), and CreatedAt. The bundle is public information (no secret key material).

#### Scenario: Bundle created with Ed25519 signing key + secp256k1 settlement key
- **WHEN** a new IdentityBundle is created
- **THEN** SigningKey.Algorithm SHALL be "ed25519" and SettlementKey.Algorithm SHALL be "secp256k1-keccak256"

#### Scenario: Bundle with ML-DSA-65 PQ signing key
- **WHEN** a new IdentityBundle is created with a PQ signing key available
- **THEN** `PQSigningKey.Algorithm` SHALL be "ml-dsa-65" and `PQGeneration` SHALL reflect the ML-DSA key derivation generation
- **AND** `BundleProofs.MLDSA65` SHALL contain an ML-DSA-65 signature over the canonical bundle bytes
- **AND** `PQSigningKey` and `PQGeneration` SHALL NOT be included in `CanonicalBundleBytes` (DID v2 hash unchanged)

#### Scenario: Bundle without PQ key (backward compat)
- **WHEN** a legacy IdentityBundle is deserialized without PQ fields
- **THEN** `PQSigningKey` SHALL be nil and `PQGeneration` SHALL be 0
- **AND** the bundle SHALL be fully functional for v1/v2 DID operations

---

### Requirement: BundleResolver interface

The system SHALL provide a `BundleResolver` interface with `ResolveBundle(did string) (*IdentityBundle, error)` for looking up remote peer IdentityBundles by DID v2 string. A `MemoryBundleCache` implementation SHALL be populated during handshakes and gossip.

#### Scenario: Bundle resolved from cache
- **WHEN** `ResolveBundle` is called with a cached DID v2
- **THEN** it SHALL return the cached IdentityBundle

#### Scenario: Unknown DID v2 returns error
- **WHEN** `ResolveBundle` is called with an uncached DID v2
- **THEN** it SHALL return an error

---

### Requirement: DIDAlias for session/reputation continuity

The system SHALL provide a `DIDAlias` type that maps v2 DID <-> v1 DID using the IdentityBundle's LegacyDID field. `CanonicalDID(did)` SHALL return the v1 DID as the canonical key for session, reputation, and firewall lookups.

#### Scenario: v2 DID resolves to v1 canonical
- **WHEN** `CanonicalDID("did:lango:v2:...")` is called and the bundle has a LegacyDID
- **THEN** it SHALL return the LegacyDID string

---

### Requirement: Identity key derivation from Master Key

The system SHALL derive the Ed25519 identity key from the Master Key using `HKDF(SHA256, MK, nil, "lango-identity-ed25519[:generation]")` where generation defaults to 0. The derivation SHALL be deterministic. The generation counter SHALL be stored in `identity-bundle.json`.

#### Scenario: Same MK produces same identity key
- **WHEN** `DeriveIdentityKey(mk, 0)` is called twice with the same MK
- **THEN** both calls SHALL return identical Ed25519 private keys

#### Scenario: PQ signing key derived from MK
- **WHEN** `DerivePQSigningKey(mk, generation)` is called
- **THEN** it SHALL derive a 32-byte seed via `HKDF-SHA256(mk, nil, "lango-pq-signing-mldsa65[:generation]")`
- **AND** produce a deterministic ML-DSA-65 keypair via `mldsa65.NewKeyFromSeed(seed)`
- **AND** the derived key SHALL be independent from the Ed25519 identity key (different HKDF domain)

#### Scenario: Same MK produces same PQ key
- **WHEN** `DerivePQSigningKey(mk, 0)` is called twice with the same MK
- **THEN** both calls SHALL return identical ML-DSA-65 private keys

---

### Requirement: Bundle file persistence

The system SHALL persist the local IdentityBundle to `~/.lango/identity-bundle.json` (0600 permissions) using atomic write (temp file + rename). Remote peer bundles SHALL be persisted to `~/.lango/known-bundles/` directory.
