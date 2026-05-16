## ADDED Requirements

### Requirement: Post-adjudication status request guards stay actionable

The post-adjudication status service SHALL preserve actionable validation errors for missing request identifiers.

#### Scenario: Missing post-adjudication transaction receipt id fails closed
- **WHEN** `postadjudicationstatus.Service.GetTransactionStatus` runs with an empty transaction receipt id
- **THEN** the call SHALL return an actionable transaction-receipt-id-required error instead of reusing the not-found path
