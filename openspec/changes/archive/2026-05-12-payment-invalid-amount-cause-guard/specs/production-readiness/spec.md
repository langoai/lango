## MODIFIED Requirements

### Requirement: Payment service has unit test coverage
The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Invalid amount error preserves the parsing cause
- **WHEN** Send is called with an invalid USDC amount string
- **THEN** the returned error SHALL still identify the request as an invalid amount
- **AND** SHALL preserve the underlying parsing cause such as malformed input or too many decimal places
