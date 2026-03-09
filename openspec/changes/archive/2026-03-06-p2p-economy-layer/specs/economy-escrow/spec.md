## ADDED Requirements

### Requirement: Escrow lifecycle state machine
The system SHALL manage escrow entries through a state machine: created → funded → (active → milestone_met)* → released | disputed | expired. Terminal states are released, disputed, and expired.

#### Scenario: Create escrow entry
- **WHEN** Create is called with payerDID, payeeDID, amount, and milestones
- **THEN** an EscrowEntry is created with Status=created, auto-generated ID, and milestone list

#### Scenario: Fund escrow
- **WHEN** Fund is called on a created escrow
- **THEN** Status transitions to "funded" and FundedAt is recorded

#### Scenario: Invalid state transition
- **WHEN** Fund is called on an already funded or released escrow
- **THEN** ErrInvalidTransition is returned

### Requirement: Milestone-based release
The system SHALL support completing milestones by index. When all milestones are completed, the escrow becomes eligible for release. Release SHALL delegate to the SettlementExecutor.

#### Scenario: Complete milestone
- **WHEN** CompleteMilestone is called with a valid milestone index on a funded escrow
- **THEN** the milestone is marked complete with a timestamp

#### Scenario: Release after all milestones
- **WHEN** Release is called and all milestones are complete
- **THEN** Status transitions to "released" and SettlementExecutor is invoked

#### Scenario: Release with incomplete milestones
- **WHEN** Release is called but milestones remain incomplete
- **THEN** ErrMilestonesIncomplete is returned

### Requirement: Dispute handling
The system SHALL allow either party to dispute a funded escrow with a reason. Disputed escrows enter a terminal state.

#### Scenario: Dispute funded escrow
- **WHEN** Dispute is called with a reason on a funded escrow
- **THEN** Status transitions to "disputed" and reason is recorded

### Requirement: Expiry check
The system SHALL expire escrow entries that exceed their configured timeout. CheckExpiry SHALL transition expired entries and return their IDs.

#### Scenario: Escrow expires
- **WHEN** CheckExpiry is called and an escrow has passed its ExpiresAt
- **THEN** Status transitions to "expired"

### Requirement: Settlement executor callback
The system SHALL use a SettlementExecutor function type to execute on-chain settlement, avoiding direct imports from the settlement package. A no-op settler SHALL be provided as default.

#### Scenario: No-op settlement
- **WHEN** Release is called with the default no-op settler
- **THEN** the release succeeds without actual on-chain transaction
