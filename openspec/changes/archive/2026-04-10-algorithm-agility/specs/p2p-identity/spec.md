## MODIFIED Requirements

### Requirement: DID Derivation from Wallet Public Key

Add a `ParseDIDPublicKey` helper that extracts raw public key bytes from a `did:lango:<hex>` string without deriving a peer ID. This function is a read-only helper for signature verification where the caller knows the key type independently. The DID format, peerID derivation, and secp256k1 requirement remain unchanged.

#### Scenario: ParseDIDPublicKey extracts bytes
- **WHEN** `ParseDIDPublicKey("did:lango:<valid-hex>")` is called
- **THEN** it SHALL return the hex-decoded byte slice without calling `peerIDFromPublicKey`

#### Scenario: ParseDIDPublicKey rejects invalid prefix
- **WHEN** `ParseDIDPublicKey` is called with a non-`did:lango:` prefix
- **THEN** it SHALL return an error

#### Scenario: ParseDIDPublicKey rejects empty key
- **WHEN** `ParseDIDPublicKey("did:lango:")` is called
- **THEN** it SHALL return an error containing "empty public key"
