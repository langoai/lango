## MODIFIED Requirements

### Requirement: Transaction submission

The payment service SHALL serialize transaction building through a nonce mutex to prevent nonce collisions. The service SHALL retry transaction submission up to 3 times with exponential backoff (1s, 2s, 4s). After successful submission, the service SHALL poll for on-chain receipt confirmation before reporting transaction status. If a post-create payment stage fails, the service SHALL attempt to mark the transaction record `failed`; if that failed-status update also fails, the service SHALL return an error that preserves both the original payment failure and the failed-status persistence failure.

#### Scenario: Concurrent payments serialize nonce acquisition
- **WHEN** multiple payments are sent concurrently
- **THEN** nonce acquisition and transaction building SHALL be serialized via mutex to prevent nonce collision

#### Scenario: Transaction submission retries
- **WHEN** `eth_sendRawTransaction` returns a transient error
- **THEN** the service SHALL retry up to 3 times with exponential backoff before failing

#### Scenario: Receipt confirmation timeout
- **WHEN** a submitted transaction does not receive a receipt within the timeout
- **THEN** the service SHALL mark the transaction as failed and return a timeout error

#### Scenario: Failed status update error is reported
- **WHEN** a payment stage fails after creating a pending transaction record
- **AND** marking that record failed also returns an error
- **THEN** `Send` SHALL return an error containing the original payment failure
- **AND** it SHALL also contain the failed-status persistence failure
