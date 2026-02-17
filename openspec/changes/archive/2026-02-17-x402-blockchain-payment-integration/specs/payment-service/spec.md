## ADDED Requirements

### Requirement: USDC ERC-20 transfer transaction building
The system SHALL build EIP-1559 ERC-20 transfer transactions by ABI-encoding `transfer(address,uint256)` with gas estimation and nonce management.

#### Scenario: Build transfer transaction
- **WHEN** `BuildTransferTx` is called with from, to, and amount
- **THEN** an EIP-1559 transaction targeting the USDC contract is returned with correct calldata

### Requirement: Payment service send flow
The system SHALL execute payments through the flow: validate address → parse amount → check spending limits → create pending record → build tx → sign → submit → update record to submitted.

#### Scenario: Successful payment
- **WHEN** `Send` is called with a valid PaymentRequest within spending limits
- **THEN** the transaction is submitted on-chain and a PaymentReceipt with txHash and status "submitted" is returned

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
The system SHALL enforce per-transaction and daily spending limits using Ent PaymentTx records. Daily totals are calculated by summing non-failed transactions since start of day.

#### Scenario: Daily spending calculated from records
- **WHEN** `DailySpent` is called
- **THEN** the sum of all pending/submitted/confirmed PaymentTx amounts for today is returned

### Requirement: USDC balance query
The system SHALL query the wallet's USDC balance via `balanceOf(address)` eth_call to the USDC contract.

#### Scenario: Query USDC balance
- **WHEN** `Balance` is called
- **THEN** the USDC balance is returned as a formatted decimal string

### Requirement: Transaction history
The system SHALL return recent PaymentTx records ordered by creation time descending.

#### Scenario: Query transaction history
- **WHEN** `History` is called with a limit
- **THEN** up to `limit` TransactionInfo records are returned, most recent first

### Requirement: PaymentTx entity schema
The system SHALL persist transaction records in an Ent PaymentTx schema with fields: id (UUID), tx_hash, from_address, to_address, amount, chain_id, status (pending/submitted/confirmed/failed), session_key, purpose, x402_url, error_message, timestamps.

#### Scenario: Failed transaction recorded
- **WHEN** a transaction fails at any step after record creation
- **THEN** the PaymentTx record is updated with status "failed" and the error message
