# session-provenance Specification

## Purpose

Durable provenance for checkpoints, session lineage, git-aware attribution, and signed provenance bundle exchange across local CLI and P2P workflows.

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

#### Scenario: Runtime child session lifecycle persists to Ent
- **WHEN** runtime multi-agent child sessions fork, merge, or discard
- **THEN** the corresponding session lineage is persisted in the `SessionProvenance` Ent store

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

### Requirement: Attribution Tracking
The system SHALL persist git-aware attribution records and join them with token usage for reporting.

#### Scenario: Workspace merge creates attribution rows
- **WHEN** a workspace branch merge completes
- **THEN** provenance attribution rows are persisted with workspace id, commit hash, file path, and line deltas

#### Scenario: Workspace bundle push creates attribution rows
- **WHEN** a signed workspace bundle is created for push or broadcast
- **THEN** provenance attribution rows are persisted for that workspace operation

#### Scenario: Non-workspace session still reports token totals
- **WHEN** a session has token usage but no workspace git evidence
- **THEN** attribution reporting still succeeds
- **AND** author/session token totals are populated

### Requirement: Provenance Configuration
The config system SHALL include a `provenance` section with: `enabled` (bool), `checkpoints.autoOnStepComplete` (bool), `checkpoints.autoOnPolicy` (bool), `checkpoints.maxPerSession` (int), `checkpoints.retentionDays` (int).

#### Scenario: Default configuration
- **WHEN** no provenance config is specified
- **THEN** defaults are: enabled=false, autoOnStepComplete=true, autoOnPolicy=true, maxPerSession=100, retentionDays=30

### Requirement: Provenance CLI
The system SHALL provide working `lango provenance` CLI commands: status, checkpoint (list|create|show), session (tree|list), attribution (show|report), and bundle (export|import).

#### Scenario: Status command
- **WHEN** `lango provenance status` is run
- **THEN** the system displays provenance configuration state

#### Scenario: Checkpoint list with filters
- **WHEN** `lango provenance checkpoint list --run <id>` is run
- **THEN** checkpoints for that run are displayed from persistent Ent store

#### Scenario: Disabled provenance message
- **WHEN** any provenance command is run with provenance.enabled=false
- **THEN** the system displays an enable instruction message

#### Scenario: Session tree command
- **WHEN** `lango provenance session tree <session-key> --depth <n>` is run
- **THEN** the CLI prints the persisted session subtree up to the requested depth

#### Scenario: Session list command
- **WHEN** `lango provenance session list --limit <n> --status <status>` is run
- **THEN** the CLI returns persisted session nodes ordered by `created_at` descending

#### Scenario: Attribution show command
- **WHEN** `lango provenance attribution show <session-key>` is run
- **THEN** the CLI returns raw attribution rows and token rollup data for the session

#### Scenario: Attribution report command
- **WHEN** `lango provenance attribution report <session-key>` is run
- **THEN** the CLI returns aggregated attribution data by author and by file

#### Scenario: Bundle export command
- **WHEN** `lango provenance bundle export <session-key> --redaction <level>` is run
- **THEN** the CLI emits a signed provenance bundle with the selected redaction level

#### Scenario: Bundle import command
- **WHEN** `lango provenance bundle import <file>` is run
- **THEN** the CLI verifies the signer DID and signature before storing provenance-owned records

#### Scenario: Remote push command
- **WHEN** `lango p2p provenance push <peer-did> <session-key> --redaction <level>` is run
- **THEN** the CLI calls the running gateway
- **AND** the gateway exports a signed bundle locally and sends it to the target peer over the provenance P2P protocol

#### Scenario: Remote fetch command
- **WHEN** `lango p2p provenance fetch <peer-did> <session-key> --redaction <level>` is run
- **THEN** the CLI calls the running gateway
- **AND** the gateway requests a signed bundle from the target peer over the provenance P2P protocol
- **AND** the returned bundle is verify-and-store imported locally

#### Scenario: Remote provenance exchange requires active session
- **WHEN** there is no active authenticated session for the target peer DID
- **THEN** remote push and fetch fail with an actionable error indicating that an active P2P session is required

### Requirement: Provenance App Module
The provenance system SHALL be registered as an appinit.Module with name "provenance", providing ProvidesProvenance, depending on ProvidesRunLedger.

#### Scenario: Module initialization
- **WHEN** the provenance module is enabled and RunLedger is available
- **THEN** the module initializes checkpoint, session tree, attribution, and bundle services

#### Scenario: Module disabled
- **WHEN** provenance.enabled is false
- **THEN** the module is skipped during app initialization

### Requirement: Checkpoint Persistence
The system SHALL persist checkpoints in the Ent database via `EntCheckpointStore` so that data survives process restarts. CLI commands and app modules MUST use `EntCheckpointStore` when `boot.DBClient` is available.

#### Scenario: Checkpoint survives process restart
- **WHEN** a checkpoint is created via CLI or auto-trigger
- **THEN** the checkpoint is persisted in the ProvenanceCheckpoint Ent table and is retrievable in subsequent CLI invocations

### Requirement: Ent Schema for Session Provenance
The system SHALL define an Ent schema `SessionProvenance` with fields: id (UUID), session_key (unique), parent_key, agent_name, goal, run_id, workspace_id, depth, status (enum), created_at, closed_at. Indexes on parent_key, agent_name, status, run_id, created_at.

#### Scenario: Session provenance schema generation
- **WHEN** `go generate ./internal/ent/...` is run
- **THEN** the SessionProvenance entity code is generated without errors

### Requirement: Ent Schema for Provenance Attribution
The system SHALL define an Ent schema `ProvenanceAttribution` with fields: id, session_key, run_id, workspace_id, author_type, author_id, file_path, commit_hash, step_id, source, lines_added, lines_removed, created_at.

#### Scenario: Attribution schema generation
- **WHEN** `go generate ./internal/ent/...` is run
- **THEN** the `ProvenanceAttribution` entity code is generated without errors

### Requirement: Provenance Bundle Verification
The system SHALL support signed provenance bundles with redaction levels `none`, `content`, and `full`.

#### Scenario: Remote peer verifies signed bundle
- **WHEN** a peer receives a provenance bundle over the provenance-specific P2P protocol
- **THEN** it verifies the bundle signature against the signer DID public key before import

#### Scenario: Tampered bundle rejected
- **WHEN** a signed bundle payload is modified after signing
- **THEN** verification fails and the bundle is rejected

### Requirement: Provenance P2P Transport
The provenance transport SHALL support both push and fetch flows.

#### Scenario: Fetch bundle request
- **WHEN** a peer receives a `fetch_bundle` provenance request with `session-key` and `redaction`
- **THEN** it exports a signed provenance bundle for that session and redaction level
- **AND** it returns the bundle over the provenance-specific P2P protocol
