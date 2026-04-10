## ADDED Requirements

### Requirement: DID v2 format

The system SHALL support a content-addressed DID format `did:lango:v2:<40-hex-chars>` where the identifier is SHA-256(canonical bundle bytes)[:20] hex-encoded. Canonical bundle bytes SHALL include Version, SigningKey, SettlementKey, and LegacyDID fields only (CreatedAt and Proofs excluded for determinism).

#### Scenario: DID v2 computed from bundle
- **WHEN** `ComputeDIDv2(bundle)` is called with an IdentityBundle
- **THEN** the result SHALL be `"did:lango:v2:" + hex(SHA256(canonical)[:20])`
- **AND** the same bundle always produces the same DID v2

### Requirement: IdentityBundle type

The system SHALL provide an `IdentityBundle` struct containing Version (int), SigningKey (PublicKeyEntry with Algorithm + PublicKey), SettlementKey (PublicKeyEntry), LegacyDID (string), Proofs (BundleProofs with Legacy + Ed25519 signatures), and CreatedAt. The bundle is public information (no secret key material).

#### Scenario: Bundle created with Ed25519 signing key + secp256k1 settlement key
- **WHEN** a new IdentityBundle is created
- **THEN** SigningKey.Algorithm SHALL be "ed25519" and SettlementKey.Algorithm SHALL be "secp256k1-keccak256"

### Requirement: BundleResolver interface

The system SHALL provide a `BundleResolver` interface with `ResolveBundle(did string) (*IdentityBundle, error)` for looking up remote peer IdentityBundles by DID v2 string. A `MemoryBundleCache` implementation SHALL be populated during handshakes and gossip.

#### Scenario: Bundle resolved from cache
- **WHEN** `ResolveBundle` is called with a cached DID v2
- **THEN** it SHALL return the cached IdentityBundle

#### Scenario: Unknown DID v2 returns error
- **WHEN** `ResolveBundle` is called with an uncached DID v2
- **THEN** it SHALL return an error

### Requirement: DIDAlias for session/reputation continuity

The system SHALL provide a `DIDAlias` type that maps v2 DID ↔ v1 DID using the IdentityBundle's LegacyDID field. `CanonicalDID(did)` SHALL return the v1 DID as the canonical key for session, reputation, and firewall lookups.

#### Scenario: v2 DID resolves to v1 canonical
- **WHEN** `CanonicalDID("did:lango:v2:...")` is called and the bundle has a LegacyDID
- **THEN** it SHALL return the LegacyDID string

### Requirement: Identity key derivation from Master Key

The system SHALL derive the Ed25519 identity key from the Master Key using `HKDF(SHA256, MK, nil, "lango-identity-ed25519[:generation]")` where generation defaults to 0. The derivation SHALL be deterministic. The generation counter SHALL be stored in `identity-bundle.json`.

#### Scenario: Same MK produces same identity key
- **WHEN** `DeriveIdentityKey(mk, 0)` is called twice with the same MK
- **THEN** both calls SHALL return identical Ed25519 private keys

### Requirement: Bundle file persistence

The system SHALL persist the local IdentityBundle to `~/.lango/identity-bundle.json` (0600 permissions) using atomic write (temp file + rename). Remote peer bundles SHALL be persisted to `~/.lango/known-bundles/` directory.

## MODIFIED Requirements

### Requirement: DID Parsing from String

`ParseDID` SHALL dispatch by prefix: `did:lango:v2:` → v2 parser, `did:lango:` → v1 parser. V2 parser returns DID{Version:2, ID:didStr} with empty PublicKey and PeerID (requires BundleResolver for full resolution). V1 parser behavior unchanged.

#### Scenario: v2 DID parsed with hollow fields
- **WHEN** `ParseDID("did:lango:v2:<hash>")` is called
- **THEN** the returned DID SHALL have Version=2, PublicKey=nil, PeerID=""

### Requirement: Peer ID Derivation from secp256k1 Public Key

`peerIDFromPublicKey` SHALL support multiple key sizes: 33-byte compressed secp256k1 via `UnmarshalSecp256k1PublicKey`, 32-byte Ed25519 via `UnmarshalEd25519PublicKey`.

#### Scenario: Ed25519 public key produces peer ID
- **WHEN** `peerIDFromPublicKey` is called with a 32-byte Ed25519 public key
- **THEN** it SHALL derive a peer ID via `crypto.UnmarshalEd25519PublicKey`
