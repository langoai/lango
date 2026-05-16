## ADDED Requirements

### Requirement: Settlement execution request guards stay actionable

The settlement execution service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing settlement execution transaction receipt id fails closed
- **WHEN** `settlementexecution.Service.ExecuteRecommendation` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable `transaction receipt id is required` error instead of panicking
