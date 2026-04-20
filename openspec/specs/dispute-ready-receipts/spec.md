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
