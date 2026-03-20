## MODIFIED Requirements

### Requirement: Append-Only Journal
The system SHALL provide an Ent-backed persistent journal implementation for RunLedger.

#### Scenario: Journal survives restart
- **WHEN** the process restarts after journal events are appended
- **THEN** the same run's journal events remain queryable

#### Scenario: App runtime prefers Ent store
- **WHEN** the shared application `ent.Client` is available during RunLedger module init
- **THEN** the module uses an Ent-backed `RunLedgerStore`
- **AND** `MemoryStore` remains a fallback for tests and non-bootstrapped contexts

### Requirement: Materialized Snapshots
The system SHALL persist cached snapshots and allow full rebuild from the journal.

#### Scenario: Snapshot cache lost
- **WHEN** a cached snapshot is missing or stale
- **THEN** the system rebuilds the snapshot by replaying the journal

#### Scenario: Snapshot and run-step projections persisted
- **WHEN** a snapshot is updated
- **THEN** the serialized snapshot is persisted in `RunSnapshot`
- **AND** the step projection rows are refreshed in `RunStep`

### Requirement: Rollout Stages
Write-through mode SHALL route workflow/background writes through RunLedger first.

#### Scenario: Write-through workflow create
- **WHEN** write-through is enabled and a workflow run is created
- **THEN** RunLedger creates the canonical `run_id` first
- **AND** workflow projection writes use that same `run_id`

#### Scenario: Projection sync failure
- **WHEN** RunLedger append succeeds but projection sync fails
- **THEN** the run remains valid in RunLedger
- **AND** the system records degraded projection state for later replay

## ADDED Requirements

### Requirement: CLI Journal Inspection
The system SHALL let operators inspect persistent RunLedger data from the CLI.

#### Scenario: `lango run list`
- **WHEN** the operator runs `lango run list`
- **THEN** the command reads recent runs from the persistent RunLedger snapshot store

#### Scenario: `lango run journal <run-id>`
- **WHEN** the operator runs `lango run journal <run-id>`
- **THEN** the command reads the persistent journal events for that run
