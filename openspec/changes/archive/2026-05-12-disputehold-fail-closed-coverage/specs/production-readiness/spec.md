## ADDED Requirements

### Requirement: Dispute hold dependency guards stay actionable

The dispute-hold service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing dispute-hold receipt store fails closed
- **WHEN** `disputehold.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing dispute-hold runtime fails closed
- **WHEN** `disputehold.Service.Execute` runs without a configured dispute-hold runtime
- **THEN** the call SHALL return an actionable `dispute hold runtime is required` error instead of panicking
