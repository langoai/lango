## MODIFIED Requirements

### Requirement: Payment service has unit test coverage

The payment service SHALL have tests covering Send error branches, History, RecordX402Payment, and failTx.

#### Scenario: Missing spending limiter fails closed
- **WHEN** `payment.Service.Send` runs without a configured spending limiter
- **THEN** the returned error SHALL still identify the failure as a spending-limit error
- **AND** SHALL preserve an actionable limiter-unavailable cause instead of panicking

#### Scenario: Missing wallet provider fails closed in send
- **WHEN** `payment.Service.Send` runs without a configured wallet provider
- **THEN** the returned error SHALL still identify the failure as a wallet-address lookup error
- **AND** SHALL preserve an actionable wallet-unavailable cause instead of panicking

#### Scenario: Missing payment store fails closed in send
- **WHEN** `payment.Service.Send` runs without a configured payment store
- **THEN** the returned error SHALL still identify the failure as a transaction-record creation error
- **AND** SHALL preserve an actionable store-unavailable cause instead of panicking

#### Scenario: Missing wallet provider fails closed in balance and info paths
- **WHEN** `payment.Service.Balance` or `payment.Service.WalletAddress` runs without a configured wallet provider
- **THEN** the returned error SHALL preserve an actionable wallet-unavailable cause instead of panicking

#### Scenario: Missing payment store fails closed in history and x402 record paths
- **WHEN** `payment.Service.History` or `payment.Service.RecordX402Payment` runs without a configured payment store
- **THEN** the returned error SHALL preserve an actionable store-unavailable cause instead of panicking
