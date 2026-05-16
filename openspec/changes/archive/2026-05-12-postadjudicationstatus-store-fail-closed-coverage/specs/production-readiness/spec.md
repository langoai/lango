## ADDED Requirements

### Requirement: Post-adjudication status dependency guards stay actionable

The post-adjudication status service SHALL preserve actionable fail-closed behavior when its backing receipt store is unavailable.

#### Scenario: Missing post-adjudication receipt store fails closed for dead-letter listing
- **WHEN** `postadjudicationstatus.Service.ListCurrentDeadLettersPage` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking

#### Scenario: Missing post-adjudication receipt store fails closed for transaction status lookup
- **WHEN** `postadjudicationstatus.Service.GetTransactionStatus` runs without a configured receipt store
- **THEN** the call SHALL return an actionable `receipt store is required` error instead of panicking
