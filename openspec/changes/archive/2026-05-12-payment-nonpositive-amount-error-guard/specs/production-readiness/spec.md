## MODIFIED Requirements

### Requirement: Payment service has unit test coverage
The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Non-positive amount keeps the business-rule error
- **WHEN** Send is called with a parsed USDC amount that is zero or negative
- **THEN** the returned error SHALL state that the amount must be positive
- **AND** SHALL NOT collapse that business-rule failure into the generic invalid-amount parse error path
