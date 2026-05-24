## MODIFIED Requirements

### Requirement: Payment service has unit test coverage
The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Wallet address failure preserves the wallet cause
- **WHEN** Send fails while retrieving the sender wallet address
- **THEN** the returned error SHALL still identify the failure as a wallet-address lookup error
- **AND** SHALL preserve the underlying wallet cause instead of collapsing into an earlier validation or spending-limit path
