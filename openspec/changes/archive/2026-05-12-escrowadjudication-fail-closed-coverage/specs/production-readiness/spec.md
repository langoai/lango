## ADDED Requirements

### Requirement: Escrow adjudication dependency guards stay actionable

The escrow adjudication service SHALL preserve actionable fail-closed behavior when its required receipt store is unavailable.

#### Scenario: Missing escrow adjudication receipt store fails closed
- **WHEN** `escrowadjudication.Service.Adjudicate` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking
