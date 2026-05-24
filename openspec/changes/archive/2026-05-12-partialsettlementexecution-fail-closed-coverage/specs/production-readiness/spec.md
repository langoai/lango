## ADDED Requirements

### Requirement: Partial settlement execution dependency guards stay actionable

The partial settlement execution service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing partial settlement receipt store fails closed
- **WHEN** `partialsettlementexecution.Service.Execute` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing partial settlement runtime fails closed
- **WHEN** `partialsettlementexecution.Service.Execute` runs without a configured direct payment runtime
- **THEN** the call SHALL return an actionable `direct payment runtime is required` error instead of panicking
