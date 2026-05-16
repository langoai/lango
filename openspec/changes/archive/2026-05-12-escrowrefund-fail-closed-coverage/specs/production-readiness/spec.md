## ADDED Requirements

### Requirement: Escrow refund dependency guards stay actionable

The escrow refund service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing escrow refund receipt store fails closed
- **WHEN** `escrowrefund.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing escrow refund runtime fails closed
- **WHEN** `escrowrefund.Service.Execute` runs without a configured refund runtime
- **THEN** the call SHALL return an actionable `escrow refund runtime is required` error instead of panicking
