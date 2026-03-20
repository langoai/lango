# RunLedger — Task OS Durable Execution Engine

## Purpose

Durable execution engine that transforms Lango from an AI chatbot into a Task OS. Provides an append-only journal as the single source of truth, typed validators via a Propose-Evidence-Verify (PEV) engine, and policy-driven failure recovery for long-running agent tasks. Core principle: "the system proves completion, not the agent."
## Requirements
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
The system SHALL support authoritative-read mode where run-state reads come from RunLedger snapshots.

#### Scenario: Authoritative snapshot read
- **WHEN** authoritative-read is enabled
- **THEN** run-state consumers read from `RunSnapshot`
- **AND** projection mirrors are no longer treated as authoritative

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
Resume SHALL be integrated with gateway/session handling while remaining opt-in.

#### Scenario: Resume candidate surfaced to user
- **WHEN** a new request expresses resume intent and a resumable paused run exists
- **THEN** the system presents resume candidates for explicit confirmation

### Requirement: Workspace Isolation
Workspace preparation SHALL be retry-safe even when the same step is validated multiple times.

#### Scenario: Repeated validation attempts
- **WHEN** the same `run_id` and `step_id` require workspace preparation more than once
- **THEN** each attempt uses a retry-safe worktree identity
- **AND** previous attempts do not cause branch-exists failures

#### Scenario: Workspace isolation gated by config
- **WHEN** `runLedger.workspaceIsolation` is `false`
- **THEN** validators still support `work_dir`
- **BUT** the app runtime does not activate `PEVEngine.WithWorkspace(...)`

#### Scenario: Workspace isolation activated
- **WHEN** `runLedger.workspaceIsolation` is `true`
- **THEN** the app runtime wires `PEVEngine.WithWorkspace(...)`
- **AND** coding-step validators execute with runtime workspace isolation enabled

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

### Requirement: Tool Governance
Each step SHALL have a `ToolProfile` that determines which tools are accessible. Profiles: `coding` (exec, fs), `browser` (browser_*), `knowledge` (search_*, rag_*), `supervisor` (run_read, run_active, run_note only). If not specified, the profile SHALL be auto-inferred from the validator type.

#### Scenario: Auto-infer coding profile
- **WHEN** a step has a `build_pass` validator and no explicit tool profile
- **THEN** the `coding` profile is assigned

#### Scenario: Auto-infer supervisor profile
- **WHEN** a step has an `orchestrator_approval` validator
- **THEN** the `supervisor` profile is assigned

### Requirement: Configuration
The system SHALL provide `RunLedgerConfig` under the root config with fields: `enabled`, `shadow`, `writeThrough`, `authoritativeRead`, `workspaceIsolation`, `staleTtl` (default: 1h), `validatorTimeout` (default: 2m), `plannerMaxRetries` (default: 2), `maxRunHistory`.

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

### Requirement: CLI Journal Inspection
The system SHALL let operators inspect persistent RunLedger data from the CLI.

#### Scenario: `lango run list`
- **WHEN** the operator runs `lango run list`
- **THEN** the command reads recent runs from the persistent RunLedger snapshot store

#### Scenario: `lango run journal <run-id>`
- **WHEN** the operator runs `lango run journal <run-id>`
- **THEN** the command reads the persistent journal events for that run

### Requirement: Command Context
The system SHALL inject active run summaries into command context.

#### Scenario: Active run summary injected
- **WHEN** an active or paused resumable run exists for the session
- **THEN** command context includes compact run summary, current blocker, and current step data

