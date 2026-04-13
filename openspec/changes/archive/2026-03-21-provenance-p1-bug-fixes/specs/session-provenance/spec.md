## MODIFIED Requirements

### Requirement: Checkpoint persistence
The system SHALL persist checkpoints in the Ent database via `EntCheckpointStore` so that data survives process restarts. CLI commands and app modules MUST use `EntCheckpointStore` when `boot.DBClient` is available.

#### Scenario: Checkpoint survives process restart
- **WHEN** a checkpoint is created via CLI or auto-trigger
- **THEN** the checkpoint is persisted in the ProvenanceCheckpoint Ent table and is retrievable in subsequent CLI invocations

#### Scenario: CLI checkpoint list returns persisted data
- **WHEN** user runs `lango provenance checkpoint list --run <id>`
- **THEN** the CLI uses `EntCheckpointStore(boot.DBClient)` and returns checkpoints from the database

### Requirement: Auto-checkpoint wiring
The `CheckpointService.OnJournalEvent` hook SHALL be registered on the RunLedger store during app module initialization via `SetAppendHook`, enabling automatic checkpoint creation on qualifying journal events.

#### Scenario: Auto-checkpoint on step validation
- **WHEN** RunLedger appends a `step_validation_passed` event and `autoOnStepComplete` is enabled
- **THEN** the hook fires and a checkpoint with trigger `step_complete` is automatically saved

### Requirement: Correct journal sequence in hooks
The `EntStore.AppendJournalEvent` SHALL assign the correct `event.Seq` value before invoking the append hook, matching the behavior of `MemoryStore`.

#### Scenario: Hook receives monotonic non-zero Seq
- **WHEN** three journal events are appended to a run via EntStore
- **THEN** the append hook receives Seq values 1, 2, 3 respectively

### Requirement: Session CLI placeholder
Session tree and list CLI subcommands SHALL display a "not yet implemented" message until a persistent session tree store is available.

#### Scenario: Session tree command
- **WHEN** user runs `lango provenance session tree <key>`
- **THEN** the CLI prints "Session tree: not yet implemented (requires persistent session tree store)"

#### Scenario: Session list command
- **WHEN** user runs `lango provenance session list`
- **THEN** the CLI prints "Session list: not yet implemented (requires persistent session tree store)"

## ADDED Requirements

### Requirement: EntCheckpointStore implements CheckpointStore
`EntCheckpointStore` SHALL implement all six methods of the `CheckpointStore` interface using the `ProvenanceCheckpoint` Ent schema. Error mapping: `ent.IsNotFound` → `ErrCheckpointNotFound`; invalid UUID → parse error.

#### Scenario: Save and retrieve checkpoint
- **WHEN** `SaveCheckpoint` is called with a valid checkpoint
- **THEN** `GetCheckpoint` with the same ID returns the checkpoint with all fields preserved

#### Scenario: Get non-existent checkpoint
- **WHEN** `GetCheckpoint` is called with a valid UUID that does not exist
- **THEN** `ErrCheckpointNotFound` is returned

#### Scenario: Delete non-existent checkpoint
- **WHEN** `DeleteCheckpoint` is called with a valid UUID that does not exist
- **THEN** `ErrCheckpointNotFound` is returned

#### Scenario: ListByRun orders by journal_seq ascending
- **WHEN** multiple checkpoints exist for a run
- **THEN** `ListByRun` returns them ordered by `journal_seq` ascending

#### Scenario: ListBySession orders by created_at descending with limit
- **WHEN** multiple checkpoints exist for a session
- **THEN** `ListBySession` returns them ordered by `created_at` descending, limited to the requested count
