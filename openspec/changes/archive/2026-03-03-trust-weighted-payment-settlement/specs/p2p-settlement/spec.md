## ADDED Requirements

### Requirement: Event-driven settlement trigger
The settlement service SHALL subscribe to `ToolExecutionPaidEvent` from the event bus and process settlements asynchronously in a separate goroutine.

#### Scenario: Paid tool execution triggers settlement
- **WHEN** a `ToolExecutionPaidEvent` is published with a valid `*eip3009.Authorization`
- **THEN** the settlement service initiates the on-chain settlement lifecycle

#### Scenario: Event with nil auth is ignored
- **WHEN** a `ToolExecutionPaidEvent` is published with nil `Auth`
- **THEN** the settlement service logs a warning and takes no action

#### Scenario: Event with wrong auth type is ignored
- **WHEN** a `ToolExecutionPaidEvent` is published with `Auth` of an unexpected type
- **THEN** the settlement service logs a warning and takes no action

### Requirement: Settlement lifecycle
The settlement service SHALL execute the full lifecycle: create DB record → build EIP-1559 transaction with `transferWithAuthorization` calldata → sign via wallet → submit with retry → wait for confirmation.

#### Scenario: Successful settlement
- **WHEN** the transaction is submitted and confirmed on-chain
- **THEN** the DB record is updated to `confirmed` status with the transaction hash

#### Scenario: Transaction submission failure with retry
- **WHEN** `SendTransaction` fails
- **THEN** the service retries up to `MaxRetries` times with exponential backoff (1s, 2s, 4s)

#### Scenario: Receipt timeout
- **WHEN** the transaction receipt is not available within `ReceiptTimeout`
- **THEN** the DB record is updated to `failed` status with timeout error

### Requirement: Nonce serialization
The settlement service SHALL serialize transaction building with a mutex to prevent nonce collisions from concurrent settlements.

#### Scenario: Concurrent settlements
- **WHEN** two settlements are triggered simultaneously
- **THEN** they are serialized and each gets a unique nonce

### Requirement: Reputation feedback on settlement outcome
The settlement service SHALL record success or failure in the reputation system after each settlement attempt.

#### Scenario: Successful settlement updates reputation
- **WHEN** settlement completes successfully
- **THEN** `RecordSuccess(peerDID)` is called on the reputation recorder

#### Scenario: Failed settlement updates reputation
- **WHEN** settlement fails (build, sign, submit, or confirmation)
- **THEN** `RecordFailure(peerDID)` is called on the reputation recorder

### Requirement: Handler publishes settlement events
The protocol handler SHALL publish `ToolExecutionPaidEvent` after successful paid tool execution when a verified authorization or deferred settlement ID is present.

#### Scenario: Verified prepay triggers event
- **WHEN** a paid tool execution succeeds with a verified EIP-3009 authorization
- **THEN** `ToolExecutionPaidEvent` is published with the authorization

#### Scenario: Post-pay triggers event with settlement ID
- **WHEN** a post-pay tool execution succeeds
- **THEN** `ToolExecutionPaidEvent` is published with the settlement ID

#### Scenario: Free tool does not trigger event
- **WHEN** a free tool execution succeeds
- **THEN** no `ToolExecutionPaidEvent` is published

### Requirement: DB record with p2p_settlement payment method
Settlement transactions SHALL be recorded in the `PaymentTx` table with `payment_method = "p2p_settlement"`.

#### Scenario: Settlement creates DB record
- **WHEN** a settlement is initiated
- **THEN** a `PaymentTx` record is created with status `pending` and method `p2p_settlement`
