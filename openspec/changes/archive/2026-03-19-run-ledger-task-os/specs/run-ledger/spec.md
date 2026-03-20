## ADDED Requirements

### Requirement: Append-Only Journal
The system SHALL record all run state changes as `JournalEvent` records with monotonic sequence numbers. Events SHALL be immutable once written.

#### Scenario: Event appended
- **WHEN** a state change occurs
- **THEN** a `JournalEvent` is appended with auto-incremented `Seq`, typed `Payload`, and `Timestamp`

### Requirement: Materialized Snapshots
The system SHALL provide `RunSnapshot` as a cached projection derived entirely from the journal. Snapshots MUST be rebuildable by replaying the journal.

#### Scenario: Cached tail replay
- **GIVEN** a cached snapshot at `LastJournalSeq = N`
- **WHEN** new events exist with `Seq > N`
- **THEN** only the tail events are replayed instead of full replay

### Requirement: PEV Engine
The system SHALL provide a Propose-Evidence-Verify engine with 6 typed validators. Custom validator types SHALL NOT be supported.

#### Scenario: orchestrator_approval never auto-passes
- **WHEN** the `orchestrator_approval` validator runs
- **THEN** it SHALL always return a failed result
- **AND** the orchestrator MUST explicitly call `run_approve_step`

### Requirement: Policy Supervisor
The orchestrator SHALL respond to step failures with one of 7 policy actions: retry, decompose, change_agent, change_validator, skip, abort, escalate.

#### Scenario: Retry policy
- **WHEN** the orchestrator applies `retry`
- **THEN** the step resets to `pending` and `RetryCount` is incremented

### Requirement: Planner Contract
The planner SHALL output strict JSON. The system SHALL validate: goal presence, step ID uniqueness, DAG acyclicity (Kahn's algorithm), valid agents, valid validator types.

#### Scenario: Dependency cycle detected
- **WHEN** step A depends on B and B depends on A
- **THEN** validation returns an error containing "cycle"

### Requirement: Run Tools
The system SHALL provide 8 agent tools with role-based access control: run_create, run_read, run_active, run_note, run_propose_step_result, run_apply_policy, run_approve_step, run_resume.

#### Scenario: Execution agent cannot complete steps
- **WHEN** an execution agent calls `run_propose_step_result`
- **THEN** the step transitions to `verify_pending`, NOT `completed`

### Requirement: Resume Protocol
Resume SHALL be opt-in only. The system SHALL detect resume intent from Korean (계속, 이어서, 마저) and English (resume, continue) keywords.

#### Scenario: Stale run excluded
- **GIVEN** a paused run last updated more than `staleTTL` ago
- **WHEN** candidates are searched
- **THEN** the stale run is NOT included

### Requirement: Workspace Isolation
Coding steps SHALL execute in git worktree isolation (fail-closed). Auto-merge SHALL be forbidden.

#### Scenario: Worktree creation failure
- **WHEN** `git worktree add` fails
- **THEN** step execution is aborted (not run on base tree)

### Requirement: Rollout Stages
The system SHALL support 4 progressive rollout stages: Shadow, Write-Through, Authoritative Read, Projection Retired.

### Requirement: Tool Governance
Each step SHALL have a `ToolProfile` (coding, browser, knowledge, supervisor) auto-inferred from validator type if not specified.

### Requirement: Configuration
The system SHALL provide `RunLedgerConfig` with: enabled, shadow, writeThrough, authoritativeRead, staleTtl, validatorTimeout, plannerMaxRetries, maxRunHistory.

### Requirement: Ent Schemas
The system SHALL provide 3 Ent schemas: RunJournal, RunSnapshot, RunStep.

### Requirement: Access Control
Tool access SHALL be role-based. Orchestrator-only: run_create, run_apply_policy, run_approve_step, run_resume. Execution-only: run_propose_step_result. All agents: run_read, run_active, run_note.
