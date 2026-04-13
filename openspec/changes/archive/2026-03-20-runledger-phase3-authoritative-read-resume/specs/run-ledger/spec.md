## MODIFIED Requirements

### Requirement: Materialized Snapshots
The system SHALL support authoritative-read mode where run-state reads come from RunLedger snapshots.

#### Scenario: Authoritative snapshot read
- **WHEN** authoritative-read is enabled
- **THEN** run-state consumers read from `RunSnapshot`
- **AND** projection mirrors are no longer treated as authoritative

### Requirement: Resume Protocol
Resume SHALL be integrated with gateway/session handling while remaining opt-in.

#### Scenario: Resume candidate surfaced to user
- **WHEN** a new request expresses resume intent and a resumable paused run exists
- **THEN** the system presents resume candidates for explicit confirmation

## ADDED Requirements

### Requirement: Command Context
The system SHALL inject active run summaries into command context.

#### Scenario: Active run summary injected
- **WHEN** an active or paused resumable run exists for the session
- **THEN** command context includes compact run summary, current blocker, and current step data
