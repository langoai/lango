## MODIFIED Requirements

### Requirement: Payment service send flow
The system SHALL execute payments through the flow: validate address → parse amount → check spending limits → create pending record → build tx → sign → submit → update record to submitted.

#### Scenario: Successful payment
- **WHEN** `Send` is called with a valid PaymentRequest within spending limits
- **THEN** the transaction is submitted on-chain and a PaymentReceipt with txHash and status "confirmed" is returned

#### Scenario: Payment exceeds per-transaction limit
- **WHEN** `Send` is called with an amount exceeding `maxPerTx`
- **THEN** an error is returned and no transaction is submitted

#### Scenario: Payment exceeds daily limit
- **WHEN** the amount plus today's total would exceed `maxDaily`
- **THEN** an error is returned and no transaction is submitted

#### Scenario: Payment to invalid address
- **WHEN** `Send` is called with an invalid Ethereum address
- **THEN** an error is returned immediately

### Requirement: Spending limits enforcement
The system SHALL enforce per-transaction and daily spending limits using a storage-facing payment usage reader. Daily totals are calculated by summing non-failed transactions since start of day.

#### Scenario: Daily spending calculated from stored records
- **WHEN** `DailySpent` is called
- **THEN** the limiter obtains usage totals through a storage-facing payment usage reader
- **AND** it does not require a direct Ent client

### Requirement: Transaction history
The system SHALL return recent PaymentTx records ordered by creation time descending.

#### Scenario: Query transaction history
- **WHEN** `History` is called with a limit
- **THEN** up to `limit` TransactionInfo records are returned, most recent first
- **AND** the history read path can be satisfied through storage-facing transaction capabilities

## ADDED Requirements

### Requirement: PaymentTx persistence abstraction
Payment transaction writes MUST flow through an explicit transaction-store interface rather than direct service-owned Ent access.

#### Scenario: Payment service records lifecycle through store interface
- **WHEN** a payment is created, submitted, confirmed, failed, or recorded as X402 activity
- **THEN** the payment service persists those transitions through a transaction-store interface
- **AND** the service does not directly access Ent-generated `PaymentTx` builders
