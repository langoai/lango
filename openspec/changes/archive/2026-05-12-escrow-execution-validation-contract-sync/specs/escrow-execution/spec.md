## ADDED Requirements

### Requirement: Escrow execution request validation stays actionable

The escrow recommendation execution path SHALL treat an empty `transaction_receipt_id` as an actionable validation error before receipt-backed rejection logic runs.

#### Scenario: Missing transaction receipt id fails validation
- **WHEN** `execute_escrow_recommendation` is called without `transaction_receipt_id`
- **THEN** the runtime SHALL return an actionable transaction-receipt-id-required error
- **AND** SHALL not pretend that receipt-backed rejection state was evaluated
