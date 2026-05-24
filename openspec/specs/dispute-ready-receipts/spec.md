## Purpose

Capability spec for dispute-ready-receipts. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Submission and transaction receipts
The system SHALL provide separate `submission receipt` and `transaction receipt` records for `knowledge exchange v1`.

#### Scenario: Submission receipt created under transaction
- **WHEN** a new artifact submission is recorded for a transaction
- **THEN** the system SHALL create a submission receipt linked to a transaction receipt

#### Scenario: Current submission pointer maintained
- **WHEN** a new submission becomes canonical for a transaction
- **THEN** the transaction receipt SHALL update its current submission pointer

### Requirement: Canonical state plus event trail
The receipt model SHALL keep canonical current state and append-only event trail separately.

#### Scenario: Canonical approval status readable
- **WHEN** a submission receipt is loaded
- **THEN** its current canonical approval status SHALL be readable directly

#### Scenario: Event trail preserved
- **WHEN** a receipt event is appended
- **THEN** the system SHALL preserve the append-only history for that submission

### Requirement: Transaction receipt payment approval state
Transaction receipts SHALL track current payment approval state for the upfront payment path.

#### Scenario: Payment approval updates transaction state
- **WHEN** an upfront payment approval outcome is applied to a transaction receipt
- **THEN** the transaction receipt SHALL update its current payment approval status

#### Scenario: Payment approval event appended
- **WHEN** an upfront payment approval outcome is applied
- **THEN** the receipt event trail SHALL append a payment approval event for later reconstruction

### Requirement: Payment execution events in receipt trails
The receipt event trail SHALL store direct payment execution authorization and denial events.

#### Scenario: Execution authorization event appended
- **WHEN** a receipt-backed direct payment execution is allowed
- **THEN** the linked receipt trail SHALL append an authorization event

#### Scenario: Execution denial event appended
- **WHEN** a receipt-backed direct payment execution is denied
- **THEN** the linked receipt trail SHALL append a denial event with reason code

### Requirement: Escrow execution state in transaction receipts
Transaction receipts SHALL track canonical escrow execution state for escrow-backed knowledge exchange transactions.

#### Scenario: Escrow recommendation input is bound
- **WHEN** an approved upfront payment decision recommends `escrow`
- **THEN** the linked transaction receipt SHALL store escrow execution input and pending escrow execution state

#### Scenario: Escrow execution reference is preserved
- **WHEN** escrow recommendation execution creates or funds an escrow
- **THEN** the transaction receipt SHALL preserve the canonical escrow reference

### Requirement: Escrow execution events in receipt trails
The receipt event trail SHALL store escrow execution progress and failure events.

#### Scenario: Escrow progress events appended
- **WHEN** receipt-backed escrow recommendation execution starts, creates, or funds an escrow
- **THEN** the linked receipt trail SHALL append the corresponding escrow execution event

#### Scenario: Escrow failure event appended
- **WHEN** receipt-backed escrow recommendation execution fails
- **THEN** the linked receipt trail SHALL append an escrow execution failure event with failure detail
