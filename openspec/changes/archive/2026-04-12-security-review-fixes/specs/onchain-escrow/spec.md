## MODIFIED Requirements

### Requirement: ZK-gated escrow prototype

#### Scenario: Verifier pinned at deployment
- **WHEN** `LangoZKEscrow` is deployed
- **THEN** the ZK verifier address SHALL be set as `immutable` in the constructor
- **AND** `releaseWithProof` SHALL NOT accept a caller-supplied verifier address
