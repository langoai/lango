## MODIFIED Requirements

### Requirement: Identity trust reputation audit documents the landed Reputation V2 contract
The `docs/architecture/identity-trust-reputation-audit.md` page SHALL describe the landed Reputation V2 contract, including separated `earnedTrustScore`, `durableNegativeUnits`, and `temporarySafetySignals`, plus the canonical trust-entry states `bootstrap`, `established`, `review`, and `temporarily_unsafe`.

#### Scenario: Audit page reflects the V2 contract
- **WHEN** a user reads `docs/architecture/identity-trust-reputation-audit.md`
- **THEN** they SHALL find the composite and earned trust distinction documented
- **AND** they SHALL find durable negative units separated from temporary safety signals
- **AND** they SHALL find the four canonical trust-entry states documented

### Requirement: P2P feature docs describe runtime trust-entry consumption
The `docs/features/p2p-network.md` and `docs/features/economy.md` pages SHALL describe how runtime consumers use the landed trust-entry contract.

#### Scenario: P2P network page describes admission and approval states
- **WHEN** a user reads `docs/features/p2p-network.md`
- **THEN** they SHALL find `bootstrap`, `established`, `review`, and `temporarily_unsafe` documented as the canonical trust-entry states
- **AND** they SHALL find `autoApproveKnownPeers` described as limited to returning peers in `established` state
- **AND** they SHALL find post-pay routing described as using earned trust for returning peers

#### Scenario: Economy page describes score consumption
- **WHEN** a user reads `docs/features/economy.md`
- **THEN** they SHALL find bootstrap peers described as using the bootstrap effective score
- **AND** they SHALL find returning peers described as using earned trust for risk and pricing inputs

### Requirement: P2P knowledge exchange track reflects landed reputation runtime integration
The `docs/architecture/p2p-knowledge-exchange-track.md` file SHALL describe the identity/trust/reputation detailed audit and the first `reputation v2 + runtime integration` slice as landed work, and it SHALL narrow the follow-on work accordingly.

#### Scenario: Track follow-on list is updated
- **WHEN** a user reads `docs/architecture/p2p-knowledge-exchange-track.md`
- **THEN** the required follow-on plan SHALL state that the identity/trust/reputation detailed audit is now landed
- **AND** they SHALL find the first `reputation v2 + runtime integration` slice described as landed work
- **AND** the remaining work SHALL be narrowed to owner-root-aware policy adoption, broader dispute-to-reputation feeds, and richer operator-facing trust/review surfaces
