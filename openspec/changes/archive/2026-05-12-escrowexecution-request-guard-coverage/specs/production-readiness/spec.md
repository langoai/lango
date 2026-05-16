## ADDED Requirements

### Requirement: Escrow execution request-id guards stay actionable

The escrow execution service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing escrow execution transaction receipt id fails closed
- **WHEN** `escrowexecution.Service.ExecuteRecommendation` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error instead of panicking
