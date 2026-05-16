## MODIFIED Requirements

### Requirement: Payment send tool
The system SHALL provide a `payment_send` tool with SafetyLevel Dangerous that accepts required `to` (address), `transaction_receipt_id`, `amount` (USDC string), and `purpose` (string) parameters. It MUST return txHash, status, amount, from, to, chainId, and network.

#### Scenario: Agent sends a receipt-backed payment
- **WHEN** the agent calls `payment_send` with `to`, `transaction_receipt_id`, `amount`, and `purpose`
- **THEN** the tool SHALL submit the payment and return a canonical payment receipt payload

### Requirement: Payment send request-id validation stays actionable

The `payment_send` tool SHALL preserve actionable validation errors for missing required receipt-linked payment inputs before direct-payment gate evaluation begins.

#### Scenario: Missing transaction receipt id returns a wrapper validation error
- **WHEN** the agent calls `payment_send` without a `transaction_receipt_id`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not defer that validation to the downstream payment gate
