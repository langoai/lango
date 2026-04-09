## Purpose

Capability spec for economy-escrow. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Escrow state machine
The system SHALL manage escrow lifecycle in `internal/economy/escrow/` through the following state machine:

```
Pending → Funded → Active → Completed → Released
                      ↓          ↓
                  Disputed    Disputed
                      ↓          ↓
                  Refunded    Refunded
                      ↓
                   Expired
```

Valid transitions:
- `Pending → Funded`: buyer deposits total amount
- `Funded → Active`: seller begins work
- `Active → Completed`: all milestones marked complete
- `Completed → Released`: funds released to seller (after dispute window)
- `Active → Disputed`: buyer or seller raises a dispute
- `Completed → Disputed`: dispute raised within DisputeWindow
- `Disputed → Refunded`: dispute resolved in buyer's favor
- `Disputed → Released`: dispute resolved in seller's favor
- `Funded → Expired`: escrow timeout reached before activation
- `Active → Expired`: escrow timeout reached during work

#### Scenario: Normal escrow flow
- **WHEN** an escrow is created, funded, work is completed, and no disputes are raised
- **THEN** the state transitions Pending → Funded → Active → Completed → Released

#### Scenario: Invalid state transition rejected
- **WHEN** a transition from "pending" to "completed" is attempted
- **THEN** an error is returned indicating the transition is not allowed

### Requirement: EscrowEntry persistence via Store interface
The system SHALL persist escrow entries through a `Store` interface with `Create`, `Get`, `List`, `ListByPeer`, `Update`, and `Delete` methods. The default implementation is `memoryStore` (in-memory with mutex protection).

#### Scenario: Create escrow entry
- **WHEN** `Store.Create(entry)` is called with a new escrow
- **THEN** the entry is stored with CreatedAt and UpdatedAt set to the current time

#### Scenario: Create duplicate escrow rejected
- **WHEN** `Store.Create(entry)` is called with an ID that already exists
- **THEN** `ErrEscrowExists` is returned

#### Scenario: List escrows by peer
- **WHEN** `Store.ListByPeer(peerDID)` is called
- **THEN** all escrows where BuyerDID or SellerDID matches peerDID are returned

### Requirement: Milestone-based release
Each escrow SHALL support multiple milestones, each with ID, Description, Amount, Status, CompletedAt, and Evidence. Funds are released proportionally as milestones are completed.

#### Scenario: Complete a milestone
- **WHEN** a milestone is marked as completed with evidence
- **THEN** MilestoneStatus changes to "completed" and CompletedAt is set

#### Scenario: All milestones completed triggers auto-release
- **WHEN** all milestones in an escrow are completed and `EscrowConfig.AutoRelease` is true
- **THEN** the escrow transitions to "completed" and then "released" after DisputeWindow

#### Scenario: Partial milestone completion
- **WHEN** 2 of 3 milestones are completed
- **THEN** `AllMilestonesCompleted()` returns false and `CompletedMilestones()` returns 2

#### Scenario: Empty milestones prevent auto-completion
- **WHEN** an escrow has zero milestones
- **THEN** `AllMilestonesCompleted()` returns false

### Requirement: Milestone status types
The system SHALL track milestone status through three states:
- `pending`: milestone not yet completed
- `completed`: milestone deliverable provided with evidence
- `disputed`: milestone outcome contested

#### Scenario: Milestone disputed
- **WHEN** a buyer disputes a milestone's completion quality
- **THEN** MilestoneStatus changes to "disputed"

### Requirement: Dispute handling
When an escrow enters "disputed" status, a `DisputeNote` SHALL be recorded on the `EscrowEntry`. Resolution results in either "refunded" (buyer wins) or "released" (seller wins).

#### Scenario: Dispute raised during active escrow
- **WHEN** a dispute is raised while escrow is "active"
- **THEN** Status transitions to "disputed" and DisputeNote is set

#### Scenario: Dispute resolved in seller's favor
- **WHEN** a dispute is resolved for the seller
- **THEN** Status transitions to "released" and funds are sent to seller

#### Scenario: Dispute resolved in buyer's favor
- **WHEN** a dispute is resolved for the buyer
- **THEN** Status transitions to "refunded" and funds are returned to buyer

### Requirement: DisputeWindow enforcement
The system SHALL enforce `EscrowConfig.DisputeWindow` (default: 1h) after completion. Disputes raised within this window are accepted; after the window closes, auto-release proceeds.

#### Scenario: Dispute within window accepted
- **WHEN** a dispute is raised within DisputeWindow after completion
- **THEN** the escrow transitions from "completed" to "disputed"

#### Scenario: Dispute after window rejected
- **WHEN** a dispute is raised after DisputeWindow has elapsed
- **THEN** the dispute is rejected and auto-release proceeds

### Requirement: Escrow expiration
Escrows SHALL expire after `EscrowConfig.DefaultTimeout` (default: 24h). Expired escrows transition to "expired" and funds are refunded to the buyer.

#### Scenario: Escrow expires during active work
- **WHEN** ExpiresAt is reached while escrow is "active"
- **THEN** Status transitions to "expired"

### Requirement: EscrowConfig defaults
The system SHALL use the following defaults from `config.EscrowConfig`:
- `Enabled`: false (opt-in)
- `DefaultTimeout`: 24h
- `MaxMilestones`: 10
- `AutoRelease`: true
- `DisputeWindow`: 1h

#### Scenario: Escrow with too many milestones rejected
- **WHEN** an escrow is created with milestones exceeding MaxMilestones (10)
- **THEN** an error is returned indicating the milestone limit

### Requirement: EscrowEntry fields
Each EscrowEntry SHALL contain: ID (UUID), BuyerDID, SellerDID, TotalAmount (*big.Int), Status, Milestones ([]Milestone), TaskID (optional), Reason, DisputeNote (optional), CreatedAt, UpdatedAt, ExpiresAt.

#### Scenario: Escrow linked to task
- **WHEN** an escrow is created for a delegated task
- **THEN** TaskID is set to the associated task identifier

### Requirement: Escrow settlement executor selection
The escrow engine SHALL use `USDCSettler` as the `SettlementExecutor` when `paymentComponents` is available (payment system enabled). The escrow engine SHALL fall back to `noopSettler` when payment is not available. The `EscrowConfig` SHALL include a `Settlement` sub-config with `ReceiptTimeout` and `MaxRetries` fields.

#### Scenario: Payment enabled uses USDC settler
- **WHEN** the economy layer is initialized with non-nil `paymentComponents`
- **THEN** `USDCSettler` is created with the payment wallet, tx builder, and RPC client

#### Scenario: Payment disabled uses noop settler
- **WHEN** the economy layer is initialized with nil `paymentComponents`
- **THEN** `noopSettler` is used and escrow operations succeed without on-chain activity

#### Scenario: Settlement config applied to settler
- **WHEN** `EscrowConfig.Settlement.ReceiptTimeout` and `MaxRetries` are configured
- **THEN** the `USDCSettler` is created with those values via functional options

#### Scenario: Released escrow triggers settlement
- **WHEN** escrow transitions to "released"
- **THEN** SettlementExecutor.Release is called to transfer TotalAmount to seller

#### Scenario: Settlement failure reverts release
- **WHEN** on-chain settlement fails
- **THEN** escrow remains in "completed" state and an error is logged

### Requirement: Store ListByStatus query
The escrow `Store` interface SHALL provide a `ListByStatus(status EscrowStatus) []*EscrowEntry` method that returns only escrows matching the given status.

#### Scenario: Query pending escrows
- **WHEN** `ListByStatus(StatusPending)` is called on a store containing escrows in pending, funded, and active statuses
- **THEN** the result SHALL contain only escrows with `Status == StatusPending`

#### Scenario: No matching escrows
- **WHEN** `ListByStatus(StatusDisputed)` is called on a store with no disputed escrows
- **THEN** the result SHALL be an empty (or nil) slice

### Requirement: Exported NoopSettler type
The escrow package SHALL export a `NoopSettler` struct that implements `SettlementExecutor` with no-op operations. All packages requiring a placeholder settler SHALL use `escrow.NoopSettler{}` instead of defining local noop types.

#### Scenario: NoopSettler satisfies interface
- **WHEN** `escrow.NoopSettler{}` is used as a `SettlementExecutor`
- **THEN** `Lock`, `Release`, and `Refund` SHALL return nil without performing any operations

#### Scenario: Compile-time interface check
- **WHEN** the escrow package is compiled
- **THEN** a `var _ SettlementExecutor = (*NoopSettler)(nil)` check SHALL verify interface compliance


### Requirement: Store ListByStatusBefore filtered query
The escrow `Store` interface SHALL provide a `ListByStatusBefore(status EscrowStatus, before time.Time) []*EscrowEntry` method that returns only escrows matching the given status AND created before the specified time.

#### Scenario: Query old pending escrows
- **WHEN** `ListByStatusBefore(StatusPending, cutoffTime)` is called
- **THEN** the result SHALL contain only escrows with `Status == StatusPending` AND `CreatedAt < cutoffTime`

#### Scenario: No matching escrows
- **WHEN** `ListByStatusBefore` is called with criteria that match no entries
- **THEN** the result SHALL be an empty (or nil) slice

#### Scenario: EntStore filters at DB level
- **WHEN** the `EntStore` implementation handles `ListByStatusBefore`
- **THEN** the query SHALL use ent predicates (`escrowdeal.Status` + `escrowdeal.CreatedAtLT`) to filter at the database level rather than loading all entries into memory

### Requirement: Hub client contract method invocation
The `HubClient` SHALL use `writeMethod` and `readMethod` helper methods to eliminate boilerplate in contract call methods. Each public method (Deposit, SubmitWork, Release, Refund, Dispute, ResolveDispute) SHALL delegate to these helpers.

#### Scenario: writeMethod wraps contract write call
- **WHEN** `Deposit(ctx, dealID)` is called
- **THEN** it delegates to `writeMethod(ctx, MethodDeposit, dealID)` which handles request construction and error wrapping

#### Scenario: readMethod wraps contract read call
- **WHEN** `NextDealID(ctx)` is called
- **THEN** it delegates to `readMethod(ctx, MethodNextDealID)` which handles request construction and error wrapping

#### Scenario: Error messages use method name
- **WHEN** a contract call fails
- **THEN** the error is wrapped as `"<methodName>: <underlying error>"` (e.g., `"deposit: rpc down"`)

### Requirement: Contract method name constants
The `hub` package SHALL define exported constants for all contract method names (V1 and V2). All client code SHALL reference these constants instead of string literals.

#### Scenario: Method constant usage
- **WHEN** `HubClient.Deposit` constructs a contract call
- **THEN** it uses `MethodDeposit` constant ("deposit") instead of a string literal

### Requirement: Transaction type constants
The escrow package SHALL define a `TransactionType` string type with constants `TxDeposit`, `TxRelease`, and `TxRefund`.

#### Scenario: Transaction type usage
- **WHEN** escrow store records a transaction
- **THEN** it uses `TxDeposit`/`TxRelease`/`TxRefund` constants instead of string literals
