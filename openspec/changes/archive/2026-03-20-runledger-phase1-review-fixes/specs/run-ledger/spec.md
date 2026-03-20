## MODIFIED Requirements

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
