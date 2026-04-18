## Purpose

Capability spec for payment-service. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: USDC ERC-20 transfer transaction building
The system SHALL build EIP-1559 ERC-20 transfer transactions by ABI-encoding `transfer(address,uint256)` with gas estimation and nonce management.

#### Scenario: Build transfer transaction
- **WHEN** `BuildTransferTx` is called with from, to, and amount
- **THEN** an EIP-1559 transaction targeting the USDC contract is returned with correct calldata

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
- **AND** the history read path can be satisfied through storage-facing transaction capabilities

### Requirement: PaymentTx entity schema
The system SHALL persist transaction records in an Ent PaymentTx schema with fields: id (UUID), tx_hash, from_address, to_address, amount, chain_id, status (pending/submitted/confirmed/failed), session_key, purpose, x402_url, error_message, timestamps.

#### Scenario: Failed transaction recorded
- **WHEN** a transaction fails at any step after record creation
- **THEN** the PaymentTx record is updated with status "failed" and the error message

### Requirement: Escrow configuration
The EscrowConfig SHALL include an `OnChain` sub-struct (`EscrowOnChainConfig`) with fields: Enabled (bool), Mode (string: "hub"|"vault"), HubAddress, VaultFactoryAddress, VaultImplementation, ArbitratorAddress, TokenAddress (all string), and PollInterval (time.Duration). All fields SHALL have `mapstructure` and `json` struct tags. The default for Enabled SHALL be false, preserving backward compatibility.

#### Scenario: On-chain config disabled by default
- **WHEN** no `economy.escrow.onChain` section is present in config
- **THEN** EscrowOnChainConfig.Enabled defaults to false and custodian mode is used

#### Scenario: Hub mode config
- **WHEN** config sets `economy.escrow.onChain.enabled=true` and `mode=hub` with `hubAddress`
- **THEN** the system initializes HubSettler with the configured hub and token addresses

### Requirement: Transaction submission
The payment service SHALL serialize transaction building through a nonce mutex to prevent nonce collisions. The service SHALL retry transaction submission up to 3 times with exponential backoff (1s, 2s, 4s). After successful submission, the service SHALL poll for on-chain receipt confirmation before reporting transaction status.

#### Scenario: Successful payment with receipt confirmation
- **WHEN** a payment is submitted and confirmed on-chain
- **THEN** the receipt status SHALL be "confirmed" with gasUsed and blockNumber populated

#### Scenario: Payment submission retry on transient failure
- **WHEN** the first SendTransaction call fails with a transient error
- **THEN** the service SHALL retry up to 3 times with exponential backoff before failing

#### Scenario: Transaction reverted on-chain
- **WHEN** a transaction is submitted but reverts on-chain (receipt.Status != 1)
- **THEN** the service SHALL return an error and update the database record to "failed"

#### Scenario: Receipt polling timeout
- **WHEN** no receipt is received within 2 minutes of submission
- **THEN** the service SHALL return a timeout error and mark the transaction as failed

#### Scenario: Concurrent payment requests
- **WHEN** two payment requests are submitted concurrently
- **THEN** nonce acquisition and transaction building SHALL be serialized via mutex to prevent nonce collision

### Requirement: PaymentReceipt fields
The PaymentReceipt struct SHALL include GasUsed (uint64) and BlockNumber (uint64) fields populated from the on-chain receipt.

#### Scenario: Receipt includes on-chain metadata
- **WHEN** a transaction is confirmed on-chain
- **THEN** the PaymentReceipt SHALL contain the actual gasUsed and blockNumber from the receipt

### Requirement: Gas fee fallback warning
The transaction builder SHALL log a WARNING when the block header's baseFee is nil and a fallback value is used.

#### Scenario: Missing baseFee in block header
- **WHEN** the block header does not contain a baseFee field
- **THEN** the builder SHALL log "WARNING: block header missing baseFee, using fallback" and use the default 1 gwei value

### Requirement: EIP-3009 signing correctness
The EIP-3009 Sign function SHALL use SignTransaction (raw signing without additional hashing) instead of SignMessage (which applies keccak256) because TypedDataHash already returns a keccak256 digest.

#### Scenario: EIP-3009 signature validity
- **WHEN** an EIP-3009 authorization is signed
- **THEN** the signature SHALL be verifiable by Verify() which uses crypto.Ecrecover on the TypedDataHash digest

#### Scenario: WalletSigner interface
- **WHEN** a wallet is used for EIP-3009 signing
- **THEN** it SHALL implement both SignTransaction (raw) and SignMessage (hashed) methods

### Requirement: PaymentTx persistence abstraction
Payment transaction writes MUST flow through an explicit transaction-store interface rather than direct service-owned Ent access.

#### Scenario: Payment service records lifecycle through store interface
- **WHEN** a payment is created, submitted, confirmed, failed, or recorded as X402 activity
- **THEN** the payment service persists those transitions through a transaction-store interface
- **AND** the service does not directly access Ent-generated `PaymentTx` builders

