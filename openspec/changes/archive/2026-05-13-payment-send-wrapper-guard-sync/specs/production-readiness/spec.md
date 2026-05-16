## ADDED Requirements

### Requirement: Payment send wrapper guards stay actionable

The `payment_send` tool SHALL preserve actionable missing-parameter errors for its required wrapper inputs before receipt-backed payment evaluation begins.

#### Scenario: Missing payment-send transaction receipt id fails at the wrapper
- **WHEN** `payment_send` is invoked without `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not defer that validation to the downstream payment gate
