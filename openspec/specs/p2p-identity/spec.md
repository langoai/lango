## ADDED Requirements

### Requirement: DID Derivation from Wallet Public Key

The `WalletDIDProvider` SHALL derive a decentralized identifier (DID) deterministically from the compressed secp256k1 public key returned by `WalletProvider.PublicKey()`. The DID format SHALL be `did:lango:<hex-encoded-compressed-pubkey>`. The derived DID SHALL be cached after the first derivation; subsequent calls to `DID()` SHALL return the cached value without calling the wallet again.

#### Scenario: DID derived on first call
- **WHEN** `WalletDIDProvider.DID(ctx)` is called for the first time
- **THEN** the provider SHALL call `wallet.PublicKey(ctx)`, construct a DID with prefix `did:lango:`, encode the public key as lowercase hex, and cache the result

#### Scenario: DID returned from cache on subsequent calls
- **WHEN** `WalletDIDProvider.DID(ctx)` is called after a successful first call
- **THEN** the provider SHALL return the cached DID without calling `wallet.PublicKey` again

#### Scenario: Wallet public key error propagates
- **WHEN** `wallet.PublicKey(ctx)` returns an error
- **THEN** `WalletDIDProvider.DID(ctx)` SHALL return a nil DID and a wrapped error; the cache SHALL NOT be populated

---

### Requirement: Peer ID Derivation from secp256k1 Public Key

The system SHALL derive a libp2p `peer.ID` from a compressed secp256k1 public key by unmarshaling it via `crypto.UnmarshalSecp256k1PublicKey` and calling `peer.IDFromPublicKey`. The derived `peer.ID` SHALL be embedded in the `DID` struct. This mapping SHALL be deterministic: the same public key always produces the same peer ID.

#### Scenario: Valid compressed public key produces peer ID
- **WHEN** `DIDFromPublicKey` is called with a valid 33-byte compressed secp256k1 public key
- **THEN** a `DID` struct SHALL be returned with a non-empty `PeerID` field derived from the key

#### Scenario: Empty public key rejected
- **WHEN** `DIDFromPublicKey` is called with an empty byte slice
- **THEN** the function SHALL return an error containing "empty public key"

#### Scenario: Invalid public key bytes rejected
- **WHEN** `DIDFromPublicKey` is called with malformed bytes that are not a valid secp256k1 point
- **THEN** the function SHALL return an error from `crypto.UnmarshalSecp256k1PublicKey`

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

### Requirement: DID Parsing from String

`ParseDID` SHALL parse a DID string in `did:lango:<hexkey>` format. It MUST validate the `did:lango:` prefix, decode the hex-encoded public key, and derive the peer ID. Any malformed input SHALL result in an error.

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
