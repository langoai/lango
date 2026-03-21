# session-provenance Specification

## Purpose
TBD - created by archiving change session-provenance. Update Purpose after archive.
## Requirements
### Requirement: Checkpoint Creation
The system SHALL support creating provenance checkpoints as thin metadata records referencing RunLedger journal positions. Checkpoints SHALL contain: ID, session_key, run_id, label, trigger type, journal_seq, optional git_ref, optional metadata, and created_at timestamp.

#### Scenario: Manual checkpoint creation
- **WHEN** a user creates a checkpoint with a label and run ID
- **THEN** the system creates a checkpoint with trigger "manual" and the current journal seq for that run

#### Scenario: Automatic checkpoint on step validation
- **WHEN** a RunLedger step validation passes and `provenance.checkpoints.autoOnStepComplete` is true
- **THEN** the system automatically creates a checkpoint with trigger "step_complete"

#### Scenario: Automatic checkpoint on policy applied
- **WHEN** a RunLedger policy decision is applied and `provenance.checkpoints.autoOnPolicy` is true
- **THEN** the system automatically creates a checkpoint with trigger "policy_applied"

#### Scenario: Max checkpoints per session enforcement
- **WHEN** a session has reached `provenance.checkpoints.maxPerSession` checkpoints
- **THEN** the system SHALL reject new checkpoint creation with ErrMaxCheckpoints

#### Scenario: Empty label rejected
- **WHEN** a checkpoint creation is attempted with an empty label
- **THEN** the system SHALL return ErrInvalidLabel

### Requirement: Checkpoint Store Interface
The system SHALL provide a CheckpointStore interface with methods: SaveCheckpoint, GetCheckpoint, ListByRun, ListBySession, CountBySession, DeleteCheckpoint. An in-memory implementation SHALL be provided for testing.

#### Scenario: List by run ordered by journal seq
- **WHEN** checkpoints are listed by run ID
- **THEN** results SHALL be ordered by journal_seq ascending

#### Scenario: List by session ordered by created_at
- **WHEN** checkpoints are listed by session key
- **THEN** results SHALL be ordered by created_at descending with optional limit

#### Scenario: Get non-existent checkpoint
- **WHEN** a checkpoint ID does not exist
- **THEN** the system SHALL return ErrCheckpointNotFound

### Requirement: RunLedger Append Hook
The RunLedger MemoryStore and EntStore SHALL accept a `WithAppendHook(func(JournalEvent))` store option. The hook SHALL be called after each successful journal event append, outside the store lock.

#### Scenario: Hook invoked after append
- **WHEN** a journal event is successfully appended and an append hook is registered
- **THEN** the hook function is called with the appended event

#### Scenario: Hook called outside lock
- **WHEN** the append hook reads from the same store
- **THEN** no deadlock occurs (hook runs after lock release)

#### Scenario: No hook registered
- **WHEN** no append hook is registered
- **THEN** journal append behavior is unchanged

### Requirement: Session Tree Tracking
The system SHALL track session hierarchy through SessionNode records containing: session_key, parent_key, agent_name, goal, run_id, workspace_id, depth, status, created_at, closed_at.

#### Scenario: Register root session
- **WHEN** a session is registered without a parent key
- **THEN** the node is created with depth 0 and status "active"

#### Scenario: Register child session
- **WHEN** a session is registered with a parent key
- **THEN** the node is created with depth = parent.depth + 1

#### Scenario: Close session
- **WHEN** a session is closed with a status (completed, merged, discarded)
- **THEN** the node's status is updated and closed_at is set

#### Scenario: Get subtree
- **WHEN** a subtree is requested for a root session with maxDepth
- **THEN** the system returns the root plus all descendants up to maxDepth levels

### Requirement: Session Lifecycle Hook
InMemoryChildStore SHALL accept a `WithLifecycleHook(func(SessionLifecycleEvent))` option. The hook SHALL be called after fork, merge, and discard operations succeed.

#### Scenario: Hook on fork
- **WHEN** a child session is forked and a lifecycle hook is registered
- **THEN** the hook is called with type "fork", child key, parent key, and agent name

#### Scenario: Hook on merge
- **WHEN** a child session is merged (with or without summary)
- **THEN** the hook is called with type "merge"

#### Scenario: Hook on discard
- **WHEN** a child session is discarded
- **THEN** the hook is called with type "discard"

### Requirement: Provenance Configuration
The config system SHALL include a `provenance` section with: `enabled` (bool), `checkpoints.autoOnStepComplete` (bool), `checkpoints.autoOnPolicy` (bool), `checkpoints.maxPerSession` (int), `checkpoints.retentionDays` (int).

#### Scenario: Default configuration
- **WHEN** no provenance config is specified
- **THEN** defaults are: enabled=false, autoOnStepComplete=true, autoOnPolicy=true, maxPerSession=100, retentionDays=30

### Requirement: Provenance CLI
The system SHALL provide `lango provenance` CLI commands: status, checkpoint (list|create|show), session (tree|list), attribution (show|report). Session tree/list and attribution show/report are not yet implemented and SHALL display placeholder messages.

#### Scenario: Status command
- **WHEN** `lango provenance status` is run
- **THEN** the system displays provenance configuration state

#### Scenario: Checkpoint list with filters
- **WHEN** `lango provenance checkpoint list --run <id>` is run
- **THEN** checkpoints for that run are displayed from persistent Ent store

#### Scenario: Disabled provenance message
- **WHEN** any provenance command is run with provenance.enabled=false
- **THEN** the system displays an enable instruction message

#### Scenario: Session commands show placeholder
- **WHEN** `lango provenance session tree` or `lango provenance session list` is run
- **THEN** the CLI prints a "not yet implemented" message

#### Scenario: Attribution commands show placeholder
- **WHEN** `lango provenance attribution show` or `lango provenance attribution report` is run
- **THEN** the CLI prints a "not yet implemented (Phase 3)" message

### Requirement: Provenance App Module
The provenance system SHALL be registered as an appinit.Module with name "provenance", providing ProvidesProvenance, depending on ProvidesRunLedger.

#### Scenario: Module initialization
- **WHEN** the provenance module is enabled and RunLedger is available
- **THEN** the module initializes CheckpointService and SessionTree with RunLedger store access

#### Scenario: Module disabled
- **WHEN** provenance.enabled is false
- **THEN** the module is skipped during app initialization

### Requirement: Ent Schema for Checkpoints
The system SHALL define an Ent schema `ProvenanceCheckpoint` with fields: id (UUID), session_key, run_id, label, trigger (enum), journal_seq, git_ref, metadata (text), created_at. Indexes on session_key, run_id, trigger, created_at, and (run_id, journal_seq).

#### Scenario: Schema generation
- **WHEN** `go generate ./internal/ent/...` is run
- **THEN** the ProvenanceCheckpoint entity code is generated without errors

### Requirement: Checkpoint Persistence
The system SHALL persist checkpoints in the Ent database via `EntCheckpointStore` so that data survives process restarts. CLI commands and app modules MUST use `EntCheckpointStore` when `boot.DBClient` is available.

#### Scenario: Checkpoint survives process restart
- **WHEN** a checkpoint is created via CLI or auto-trigger
- **THEN** the checkpoint is persisted in the ProvenanceCheckpoint Ent table and is retrievable in subsequent CLI invocations

#### Scenario: CLI checkpoint list returns persisted data
- **WHEN** user runs `lango provenance checkpoint list --run <id>`
- **THEN** the CLI uses `EntCheckpointStore(boot.DBClient)` and returns checkpoints from the database

### Requirement: Auto-checkpoint Wiring
The `CheckpointService.OnJournalEvent` hook SHALL be registered on the RunLedger store during app module initialization via `SetAppendHook`, enabling automatic checkpoint creation on qualifying journal events.

#### Scenario: Auto-checkpoint on step validation
- **WHEN** RunLedger appends a `step_validation_passed` event and `autoOnStepComplete` is enabled
- **THEN** the hook fires and a checkpoint with trigger `step_complete` is automatically saved

### Requirement: Correct Journal Sequence in Hooks
The `EntStore.AppendJournalEvent` SHALL assign the correct `event.Seq` value before invoking the append hook, matching the behavior of `MemoryStore`.

#### Scenario: Hook receives monotonic non-zero Seq
- **WHEN** three journal events are appended to a run via EntStore
- **THEN** the append hook receives Seq values 1, 2, 3 respectively

### Requirement: Session CLI Placeholder
Session tree and list CLI subcommands SHALL display a "not yet implemented" message until a persistent session tree store is available.

#### Scenario: Session tree command
- **WHEN** user runs `lango provenance session tree <key>`
- **THEN** the CLI prints "Session tree: not yet implemented (requires persistent session tree store)"

#### Scenario: Session list command
- **WHEN** user runs `lango provenance session list`
- **THEN** the CLI prints "Session list: not yet implemented (requires persistent session tree store)"

### Requirement: EntCheckpointStore Implements CheckpointStore
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

### Requirement: Ent Schema for Session Provenance
The system SHALL define an Ent schema `SessionProvenance` with fields: id (UUID), session_key (unique), parent_key, agent_name, goal, run_id, workspace_id, depth, status (enum), created_at, closed_at. Indexes on parent_key, agent_name, status, run_id, created_at.

#### Scenario: Schema generation
- **WHEN** `go generate ./internal/ent/...` is run
- **THEN** the SessionProvenance entity code is generated without errors

