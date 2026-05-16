## ADDED Requirements

### Requirement: Settlement execution dependency guards stay actionable

The direct settlement execution service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing settlement execution receipt store fails closed
- **WHEN** `settlementexecution.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing settlement execution runtime fails closed
- **WHEN** `settlementexecution.Service.Execute` runs without a configured direct payment runtime
- **THEN** the call SHALL return an actionable `direct payment runtime is required` error instead of panicking
