## ADDED Requirements

### Requirement: Post-adjudication replay request guards stay actionable

The post-adjudication replay service SHALL preserve actionable validation errors for missing replay request identifiers.

#### Scenario: Missing replay transaction receipt id fails closed
- **WHEN** `postadjudicationreplay.Service.Replay` runs with an empty `transaction_receipt_id`
- **THEN** the call SHALL return an actionable `transaction_receipt_id is required` error instead of panicking
