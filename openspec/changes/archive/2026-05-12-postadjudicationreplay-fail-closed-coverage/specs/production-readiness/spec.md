## ADDED Requirements

### Requirement: Post-adjudication replay dependency guards stay actionable

The post-adjudication replay service SHALL preserve actionable fail-closed behavior when its required runtime dependencies are unavailable.

#### Scenario: Missing replay receipt store fails closed
- **WHEN** `postadjudicationreplay.Service.Replay` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing replay dispatcher fails closed
- **WHEN** `postadjudicationreplay.Service.Replay` runs without a configured dispatcher
- **THEN** the call SHALL return an actionable `dispatcher is required` error instead of panicking
