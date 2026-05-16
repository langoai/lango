## MODIFIED Requirements

### Requirement: Payment service has unit test coverage

The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Missing transaction RPC fails closed during submission
- **WHEN** `payment.Service.submitWithRetry` runs without a configured transaction RPC client
- **THEN** the returned error SHALL preserve an actionable transaction-RPC-unavailable cause instead of panicking

#### Scenario: Missing transaction RPC fails closed during confirmation
- **WHEN** `payment.Service.waitForConfirmation` runs without a configured transaction RPC client
- **THEN** the returned error SHALL preserve an actionable receipt-RPC-unavailable cause instead of panicking
