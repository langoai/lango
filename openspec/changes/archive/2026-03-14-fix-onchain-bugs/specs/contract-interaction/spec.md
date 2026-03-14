## MODIFIED Requirements

### Requirement: Write transaction receipt handling
The contract caller's Write() method SHALL check the receipt status after waiting for the transaction receipt. If the receipt status is not successful (ReceiptStatusSuccessful), Write() SHALL return an ErrTxReverted error with the transaction hash and status. If the receipt times out, Write() SHALL return an ErrReceiptTimeout error instead of silently returning a partial result.

#### Scenario: Transaction reverts on-chain
- **WHEN** a Write() call submits a transaction that gets mined but reverts
- **THEN** Write() returns an error wrapping ErrTxReverted with the tx hash and receipt status

#### Scenario: Receipt timeout
- **WHEN** a Write() call submits a transaction but the receipt is not available within the timeout period
- **THEN** Write() returns an error wrapping ErrReceiptTimeout with the tx hash

#### Scenario: Successful transaction
- **WHEN** a Write() call submits a transaction that gets mined with status=1 (success)
- **THEN** Write() returns a ContractCallResult with TxHash and GasUsed, and nil error
