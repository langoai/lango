## MODIFIED Requirements

### Requirement: Payment service has unit test coverage
The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Spending limit failure preserves limiter cause
- **WHEN** Send is called with an amount that the spending limiter rejects
- **THEN** the returned error SHALL still identify the failure as a spending-limit error
- **AND** SHALL preserve the underlying limiter cause instead of collapsing into an earlier validation path
