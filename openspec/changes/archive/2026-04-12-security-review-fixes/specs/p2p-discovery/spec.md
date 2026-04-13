## MODIFIED Requirements

### Requirement: GossipCard signing

#### Scenario: Signed gossip card verified on receive
- **WHEN** a gossip card with signature fields is received
- **THEN** `handleMessage` SHALL call `VerifyCardSignature` before storing the card
- **AND** cards with invalid signatures SHALL be discarded with a warning log

#### Scenario: Empty bundle in signed card rejected
- **WHEN** a signed gossip card has a non-empty `Bundle` field but the bundle contains no valid `SigningKey.PublicKey`
- **THEN** the card SHALL be rejected with "signed card has bundle but no valid signing key"

#### Scenario: Card DID matches bundle v2 DID only
- **WHEN** a signed gossip card with a bundle is verified
- **THEN** `card.DID` SHALL match `ComputeDIDv2(bundle)` only
- **AND** `bundle.LegacyDID` SHALL NOT be accepted as a valid match (unverifiable without `Proofs.Legacy`)

#### Scenario: Bundle-less signed card accepted for backward compat
- **WHEN** a signed gossip card has no `Bundle` field (pre-upgrade peer)
- **THEN** classical and PQ signature verification SHALL be skipped
- **AND** the card SHALL be accepted for backward compatibility

#### Scenario: SignatureAlgorithm set before canonical payload
- **WHEN** `signCard` computes the canonical payload for signing
- **THEN** `SignatureAlgorithm` SHALL be set before `CanonicalCardPayload` is called
- **AND** sender and receiver SHALL hash the same JSON fields
