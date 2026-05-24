## MODIFIED Requirements

### Requirement: Payment service has unit test coverage

The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Missing balance builder fails closed
- **WHEN** `payment.Service.Balance` runs without a configured transaction builder
- **THEN** the returned error SHALL still identify the failure as a balance-query error
- **AND** SHALL preserve an actionable builder-unavailable cause instead of panicking

#### Scenario: Missing balance RPC client fails closed
- **WHEN** `payment.Service.Balance` runs without a configured RPC client
- **THEN** the returned error SHALL still identify the failure as a balance-query error
- **AND** SHALL preserve an actionable balance-RPC-unavailable cause instead of panicking
