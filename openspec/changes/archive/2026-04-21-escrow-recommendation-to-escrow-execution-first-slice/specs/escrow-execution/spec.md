## ADDED Requirements

### Requirement: Receipt-backed escrow recommendation execution
The system SHALL provide a receipt-backed escrow recommendation execution path for `knowledge exchange v1`. The first slice SHALL execute only `create + fund`.

#### Scenario: Approved escrow recommendation executes
- **WHEN** `execute_escrow_recommendation` is called for a transaction receipt whose canonical payment approval state is approved, whose canonical settlement hint is `escrow`, and whose escrow execution input is already bound
- **THEN** the runtime SHALL create and fund the escrow and return the transaction receipt ID, submission receipt ID, escrow reference, and escrow execution status

#### Scenario: Current submission is resolved from transaction receipt
- **WHEN** `execute_escrow_recommendation` is called with a valid `transaction_receipt_id`
- **THEN** the runtime SHALL resolve the current canonical submission from the linked transaction receipt instead of requiring a separate submission ID input

#### Scenario: Unsupported receipt state denies execution
- **WHEN** `execute_escrow_recommendation` is called for a transaction receipt that is missing approval, missing escrow input, missing a current submission, or already progressed beyond the initial pending state
- **THEN** the runtime SHALL reject execution

### Requirement: Escrow execution evidence
The system SHALL persist canonical escrow execution state and append-only escrow execution events onto receipts.

#### Scenario: Execution progress updates transaction receipt
- **WHEN** escrow recommendation execution progresses through the first slice
- **THEN** the transaction receipt SHALL track escrow execution status, escrow reference, and bound escrow execution input

#### Scenario: Execution failure is preserved
- **WHEN** escrow creation or funding fails
- **THEN** the receipt trail SHALL append an escrow execution failure event and preserve the failure reason for later reconstruction
