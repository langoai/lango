## ADDED Requirements

### Requirement: Escrow release dependency guards stay actionable

The escrow release service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing escrow release receipt store fails closed
- **WHEN** `escrowrelease.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing escrow release runtime fails closed
- **WHEN** `escrowrelease.Service.Execute` runs without a configured escrow runtime
- **THEN** the call SHALL return an actionable `escrow runtime is required` error instead of panicking
