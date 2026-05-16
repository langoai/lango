## MODIFIED Requirements

### Requirement: Payment service has unit test coverage

The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Missing transaction builder fails closed after record creation
- **WHEN** `payment.Service.Send` reaches transaction construction with no configured `TxBuilder`
- **THEN** the returned error SHALL still identify the failure as a build-transaction error
- **AND** SHALL preserve an actionable builder-unavailable cause instead of panicking the send path

#### Scenario: Missing transaction RPC client fails closed after record creation
- **WHEN** `payment.Service.Send` reaches transaction construction with a `TxBuilder` that has no RPC client
- **THEN** the returned error SHALL still identify the failure as a build-transaction error
- **AND** SHALL preserve an actionable transaction-RPC-unavailable cause instead of panicking the send path
