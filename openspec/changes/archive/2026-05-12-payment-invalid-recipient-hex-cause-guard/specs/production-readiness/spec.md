## MODIFIED Requirements

### Requirement: Payment service has unit test coverage
The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Invalid recipient error preserves format or hex validation cause
- **WHEN** Send is called with an invalid Ethereum address
- **THEN** the returned error SHALL still identify the request as an invalid recipient
- **AND** SHALL preserve the underlying validation cause whether the address fails format validation or hex validation
