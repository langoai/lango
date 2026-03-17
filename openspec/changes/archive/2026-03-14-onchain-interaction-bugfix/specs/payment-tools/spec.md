## MODIFIED Requirements

### Requirement: Payment send tool status reporting
The payment_send tool SHALL return the actual on-chain status from the receipt instead of a hardcoded "submitted" string. When available, the tool SHALL include gasUsed and blockNumber in the response.

#### Scenario: Confirmed payment response
- **WHEN** a payment is confirmed on-chain
- **THEN** the tool response SHALL include `status: "confirmed"`, `gasUsed`, and `blockNumber` from the receipt

#### Scenario: Failed payment response
- **WHEN** a payment transaction reverts on-chain
- **THEN** the tool SHALL return an error (not a success response with status "submitted")

#### Scenario: Response includes gas metadata
- **WHEN** gasUsed > 0 in the receipt
- **THEN** the tool response SHALL include the gasUsed field
