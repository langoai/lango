# RunLedger — Task OS Durable Execution Engine

## Purpose

Durable execution engine that transforms Lango from an AI chatbot into a Task OS. Provides an append-only journal as the single source of truth, typed validators via a Propose-Evidence-Verify (PEV) engine, and policy-driven failure recovery for long-running agent tasks. Core principle: "the system proves completion, not the agent."
## Requirements
### Requirement: Append-Only Journal
The system SHALL record all run state changes as `JournalEvent` records with monotonic sequence numbers. Events SHALL be immutable once written — no overwrite, no delete, only append.

#### Scenario: Event appended
- **WHEN** a state change occurs (e.g., step started, result proposed)
- **THEN** a `JournalEvent` is appended with auto-incremented `Seq` within the run
- **AND** the event includes a typed `Payload` (JSON) and a `Timestamp`

#### Scenario: Event types
- **GIVEN** the 14 event types: run_created, plan_attached, step_started, step_result_proposed, step_validation_passed, step_validation_failed, policy_decision_applied, note_written, run_paused, run_resumed, run_completed, run_failed, projection_synced, criterion_met
- **WHEN** any run lifecycle transition occurs
- **THEN** the corresponding event type is recorded

### Requirement: Materialized Snapshots
The system SHALL provide `RunSnapshot` as a cached projection derived entirely from the journal. Snapshots SHALL never be the source of truth — they MUST be rebuildable by replaying the journal.

#### Scenario: Full materialization
- **WHEN** `MaterializeFromJournal(events)` is called with a complete event list
- **THEN** a `RunSnapshot` is produced reflecting the current state of the run

#### Scenario: Cached tail replay
- **GIVEN** a cached snapshot at `LastJournalSeq = N`
- **WHEN** new events exist with `Seq > N`
- **THEN** only the tail events are replayed via `ApplyTail` instead of full replay
- **AND** the cached snapshot is updated with the new `LastJournalSeq`

#### Scenario: Empty journal
- **WHEN** `MaterializeFromJournal` is called with an empty event list
- **THEN** an error is returned

### Requirement: Run Lifecycle
A Run SHALL transition through statuses: `planning` → `running` → `paused` | `completed` | `failed`. Status transitions SHALL occur only through journal events.

#### Scenario: Run created
- **WHEN** `EventRunCreated` is recorded
- **THEN** the run status is set to `planning`

#### Scenario: Plan attached
- **WHEN** `EventPlanAttached` is recorded with steps and acceptance criteria
- **THEN** the run status transitions to `running`

#### Scenario: Run paused
- **WHEN** `EventRunPaused` is recorded (e.g., turn limit reached)
- **THEN** the run status transitions to `paused`

#### Scenario: Run completed
- **WHEN** all steps are terminal AND all acceptance criteria are met
- **THEN** `EventRunCompleted` is recorded and status transitions to `completed`

#### Scenario: Run completion check after step verification
- **WHEN** a step verification passes and `checkRunCompletion` is invoked
- **THEN** if all steps are successful: acceptance criteria are verified, `EventCriterionMet` is journaled for each satisfied criterion, and run transitions to `completed` or `failed`
- **AND** if all steps are terminal but NOT all successful: run transitions to `failed`
- **AND** if steps are still running: run remains in `running` status

### Requirement: Step Lifecycle
Each step SHALL transition: `pending` → `in_progress` → `verify_pending` → `completed` | `failed` | `interrupted`. Execution agents MUST NOT directly change step status to `completed` — only the PEV engine MAY do so.

#### Scenario: Step started
- **WHEN** `EventStepStarted` is recorded for a step
- **THEN** step status transitions to `in_progress`

#### Scenario: Result proposed
- **WHEN** an execution agent calls `run_propose_step_result`
- **THEN** `EventStepResultProposed` is recorded
- **AND** step status transitions to `verify_pending`
- **AND** the step is NOT marked as completed

#### Scenario: Validation passed
- **WHEN** the PEV engine records `EventStepValidationPassed`
- **THEN** step status transitions to `completed`

#### Scenario: Validation failed
- **WHEN** the PEV engine records `EventStepValidationFailed`
- **THEN** step status transitions to `failed`
- **AND** a `PolicyRequest` is generated for the orchestrator

### Requirement: Dependency Resolution
Steps SHALL declare dependencies via `DependsOn` (list of step IDs). A step MUST NOT start until all its dependencies have status `completed`.

#### Scenario: Next executable step
- **GIVEN** step A is completed, step B depends on A, step C depends on B
- **WHEN** the system selects the next executable step
- **THEN** step B is returned (C is not ready because B is not completed)

#### Scenario: No step ready
- **GIVEN** all pending steps have unmet dependencies
- **WHEN** the system selects the next executable step
- **THEN** nil is returned

### Requirement: PEV Engine
The system SHALL provide a Propose-Evidence-Verify engine that runs typed validators against step results. The PEV engine SHALL record validation results in the journal. It SHALL NOT modify step status directly — status changes happen via journal event replay.

#### Scenario: Validator passes
- **WHEN** the PEV engine runs a validator and it passes
- **THEN** `EventStepValidationPassed` is recorded
- **AND** no `PolicyRequest` is returned

#### Scenario: Validator fails
- **WHEN** the PEV engine runs a validator and it fails
- **THEN** `EventStepValidationFailed` is recorded
- **AND** a `PolicyRequest` is returned with failure details, retry count, and max retries

#### Scenario: Unknown validator type
- **WHEN** a step references an unregistered validator type
- **THEN** an error is returned

#### Scenario: Auto-verification on propose
- **WHEN** an execution agent calls `run_propose_step_result`
- **THEN** `EventStepResultProposed` is recorded
- **AND** the PEV engine automatically runs the registered validator for the step — no manual trigger needed
- **AND** on pass: `EventStepValidationPassed` is recorded, step transitions to `completed`, and run completion is checked
- **AND** on fail: a structured payload is returned containing `failure_reason` for the orchestrator's policy decision

### Requirement: Typed Validators
The system SHALL provide 6 built-in validators. Custom validator types SHALL NOT be supported to prevent auto-pass.

#### Scenario: build_pass validator
- **WHEN** the `build_pass` validator runs
- **THEN** it executes `go build <target>` and reports pass/fail based on exit code

#### Scenario: test_pass validator
- **WHEN** the `test_pass` validator runs
- **THEN** it executes `go test <target>` and reports pass/fail based on exit code

#### Scenario: file_changed validator
- **WHEN** the `file_changed` validator runs with a target pattern
- **THEN** it checks `git diff --name-only HEAD` for matching files

#### Scenario: artifact_exists validator
- **WHEN** the `artifact_exists` validator runs with a target path
- **THEN** it checks `os.Stat(<target>)` for file existence

#### Scenario: command_pass validator
- **WHEN** the `command_pass` validator runs
- **THEN** it executes the target command and checks exit code against `expected_exit_code` (default: 0)

#### Scenario: orchestrator_approval validator
- **WHEN** the `orchestrator_approval` validator runs
- **THEN** it SHALL always return a failed result ("awaiting orchestrator approval")
- **AND** the orchestrator MUST explicitly call `run_approve_step` to pass

#### Scenario: WorkDir injection
- **GIVEN** `ValidatorSpec` includes a `work_dir` field set at runtime by the workspace manager
- **WHEN** a command-running validator (`build_pass`, `test_pass`, `file_changed`, `command_pass`) executes
- **THEN** the command's working directory is set via `cmd.Dir = spec.WorkDir`
- **AND** the `artifact_exists` validator resolves paths via `filepath.Join(spec.WorkDir, target)`
- **AND** in Phase 1, `WorkDir` is empty (no isolation — commands run in the default directory)
- **AND** in Phase 3, `pev.WithWorkspace()` activates full worktree isolation with a populated `WorkDir`

### Requirement: Policy Supervisor
The orchestrator SHALL respond to `PolicyRequest` with one of 7 actions: `retry`, `decompose`, `change_agent`, `change_validator`, `skip`, `abort`, `escalate`. The decision is recorded as `EventPolicyDecisionApplied`.

#### Scenario: Retry policy
- **WHEN** the orchestrator applies `retry`
- **THEN** the step status resets to `pending` and `RetryCount` is incremented

#### Scenario: Decompose policy
- **WHEN** the orchestrator applies `decompose` with new sub-steps
- **THEN** the original step is marked completed
- **AND** new steps are appended to the run

#### Scenario: Abort policy
- **WHEN** the orchestrator applies `abort`
- **THEN** the run status transitions to `failed`

#### Scenario: Escalate policy
- **WHEN** the orchestrator applies `escalate`
- **THEN** the run's `CurrentBlocker` is set to "escalated: <reason>"

### Requirement: Planner Contract
The planner SHALL output strict JSON (optionally in ````json` fences). The system SHALL validate: goal presence, step ID uniqueness, dependency DAG acyclicity (Kahn's algorithm), valid agent names, and valid validator types.

#### Scenario: Valid plan parsed
- **WHEN** the planner outputs a JSON plan with goal, steps, and acceptance_criteria
- **THEN** `ParsePlannerOutput` successfully deserializes it
- **AND** `ValidatePlanSchema` passes

#### Scenario: Fenced JSON parsed
- **WHEN** the planner wraps JSON in ````json ... ``` `` fences
- **THEN** the JSON is extracted and parsed successfully

#### Scenario: Dependency cycle detected
- **WHEN** step A depends on B and step B depends on A
- **THEN** `ValidatePlanSchema` returns an error containing "cycle"

#### Scenario: Unknown agent rejected
- **WHEN** a step references an agent not in `validAgents`
- **THEN** `ValidatePlanSchema` returns an error containing "unknown agent"

#### Scenario: Parse failure
- **WHEN** the planner output contains invalid JSON
- **THEN** `ParsePlannerOutput` returns `ErrInvalidPlanJSON`

### Requirement: Run Tools
The system SHALL validate execution-agent step proposals before journaling them.

#### Scenario: Authorized proposal
- **WHEN** an execution agent calls `run_propose_step_result` for its own step
- **AND** the step exists
- **AND** the step status is `in_progress`
- **THEN** `EventStepResultProposed` is appended
- **AND** auto-verification proceeds

#### Scenario: Unknown step rejected before journaling
- **WHEN** an execution agent calls `run_propose_step_result` for a nonexistent step
- **THEN** an error is returned
- **AND** no `EventStepResultProposed` is appended

#### Scenario: Wrong owner rejected before journaling
- **WHEN** an execution agent calls `run_propose_step_result` for a step owned by a different agent
- **THEN** `ErrAccessDenied` is returned
- **AND** no `EventStepResultProposed` is appended

#### Scenario: Wrong pre-state rejected before journaling
- **WHEN** an execution agent calls `run_propose_step_result` for a step not in `in_progress`
- **THEN** an error is returned
- **AND** no `EventStepResultProposed` is appended

### Requirement: Resume Protocol
Resume SHALL be opt-in only — no automatic resurrection. The system SHALL detect resume intent from user messages (Korean: 계속, 이어서, 마저; English: resume, continue) and present candidates for explicit confirmation.

#### Scenario: Resume intent detected
- **WHEN** a user message contains "계속해줘" or "resume the task"
- **THEN** `DetectResumeIntent` returns true

#### Scenario: No resume intent
- **WHEN** a user message contains "build a new feature"
- **THEN** `DetectResumeIntent` returns false

#### Scenario: Stale run excluded
- **GIVEN** a paused run last updated more than `staleTTL` (default: 1h) ago
- **WHEN** resume candidates are searched
- **THEN** the stale run is NOT included

### Requirement: Workspace Isolation
Workspace preparation SHALL be retry-safe even when the same step is validated multiple times.

#### Scenario: Repeated validation attempts
- **WHEN** the same `run_id` and `step_id` require workspace preparation more than once
- **THEN** each attempt uses a retry-safe worktree identity
- **AND** previous attempts do not cause branch-exists failures

#### Scenario: Phase 1 runtime readiness only
- **WHEN** RunLedger is enabled in the current Phase 1 runtime
- **THEN** validators support `work_dir`
- **BUT** the app runtime does not yet activate `PEVEngine.WithWorkspace(...)`
- **AND** full workspace isolation activation remains part of the later execution-isolation phase

### Requirement: Rollout Stages
The system SHALL support 4 progressive rollout stages: Shadow (journal only), Write-Through (ledger first, then mirror), Authoritative Read (reads from ledger), Projection Retired (legacy removed).

#### Scenario: Shadow mode
- **GIVEN** `runLedger.shadow: true`
- **WHEN** runs are created
- **THEN** journal events are recorded but existing workflow/background systems operate unchanged

### Requirement: Tool Governance
Each step SHALL have a `ToolProfile` that determines which tools are accessible. Profiles: `coding` (exec, fs), `browser` (browser_*), `knowledge` (search_*, rag_*), `supervisor` (run_read, run_active, run_note only). If not specified, the profile SHALL be auto-inferred from the validator type.

#### Scenario: Auto-infer coding profile
- **WHEN** a step has a `build_pass` validator and no explicit tool profile
- **THEN** the `coding` profile is assigned

#### Scenario: Auto-infer supervisor profile
- **WHEN** a step has an `orchestrator_approval` validator
- **THEN** the `supervisor` profile is assigned

### Requirement: Configuration
The system SHALL provide `RunLedgerConfig` under the root config with fields: `enabled`, `shadow`, `writeThrough`, `authoritativeRead`, `staleTtl` (default: 1h), `validatorTimeout` (default: 2m), `plannerMaxRetries` (default: 2), `maxRunHistory`.

#### Scenario: Default config
- **WHEN** no RunLedger config is provided
- **THEN** the system defaults to disabled

#### Scenario: Enabled with shadow
- **WHEN** `runLedger.enabled: true` and `runLedger.shadow: true`
- **THEN** the RunLedger module initializes in shadow mode

### Requirement: Ent Schemas
The system SHALL provide 3 Ent schemas: `RunJournal` (append-only event log with run_id+seq unique index), `RunSnapshot` (cached materialized view with unique run_id), `RunStep` (step projection with run_id+step_id unique index).

#### Scenario: Journal uniqueness
- **GIVEN** the RunJournal schema
- **WHEN** two events with the same run_id and seq are inserted
- **THEN** a unique constraint violation occurs

### Requirement: Access Control
Tool access SHALL be role-based. The orchestrator (agent name "orchestrator" or "lango-orchestrator") MAY call `run_create`, `run_apply_policy`, `run_approve_step`, `run_resume`. Execution agents MAY call `run_propose_step_result`. All agents MAY call `run_read`, `run_active`, `run_note`.

#### Scenario: Execution agent blocked from run_create
- **WHEN** a non-orchestrator agent calls `run_create`
- **THEN** `ErrAccessDenied` is returned

#### Scenario: Orchestrator blocked from execution-only tools
- **WHEN** the orchestrator calls `run_propose_step_result`
- **THEN** `ErrAccessDenied` is returned
- **AND** only execution agents MAY call execution-only tools

#### Scenario: run_approve_step restricted to orchestrator_approval steps
- **WHEN** the orchestrator calls `run_approve_step` for a step
- **THEN** the step MUST have the `orchestrator_approval` validator type
- **AND** the step MUST be in `verify_pending` or `failed` status
- **AND** if either condition is not met, an error is returned

