## MODIFIED Requirements

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
