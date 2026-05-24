## ADDED Requirements

### Requirement: Payment send request-id validation stays actionable

The `payment_send` tool SHALL preserve actionable direct-payment gate validation errors for missing transaction receipt ids instead of translating them into denied business results.

#### Scenario: Missing transaction receipt id returns a validation error
- **WHEN** the agent calls `payment_send` without a `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable transaction-receipt-id-required error
- **AND** SHALL not synthesize a denied `missing_receipt` result
