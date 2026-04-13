## ADDED Requirements

### Requirement: GossipCard signing

GossipCards SHALL support dual signatures (classical + PQ). The `CanonicalCardPayload()` function SHALL serialize all card fields except `Signature` and `PQSignature`. The `Bundle` and `PQSignerPublicKey` fields SHALL be included in the canonical payload to prevent bundle substitution. Unsigned legacy cards SHALL be accepted for backward compatibility.

#### Scenario: Signed gossip card published
- **WHEN** a gossip card is published with a signer available
- **THEN** `Signature` and `SignatureAlgorithm` SHALL be set from the classical signer
- **AND** if a PQ signer is available, `PQSignature`, `PQSignatureAlgorithm`, and `PQSignerPublicKey` SHALL be set

#### Scenario: Signed gossip card verified on receive
- **WHEN** a gossip card with signature fields is received
- **THEN** the classical signature SHALL be verified against the canonical payload
- **AND** if PQ signature is present, it SHALL be verified against the embedded `PQSignerPublicKey`

#### Scenario: Unsigned legacy card accepted
- **WHEN** a gossip card without signature fields is received
- **THEN** the card SHALL be accepted for backward compatibility

#### Scenario: Tampered card rejected
- **WHEN** a signed gossip card with a modified field is received
- **THEN** signature verification SHALL fail and the card SHALL be discarded

#### Scenario: Bundle included in signed payload
- **WHEN** `CanonicalCardPayload()` is computed
- **THEN** the Bundle field SHALL be included in the canonical payload
- **AND** an attacker cannot substitute a different Bundle without invalidating the signature
