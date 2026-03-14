## MODIFIED Requirements

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

## ADDED Requirements

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
