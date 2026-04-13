## MODIFIED Requirements

### Requirement: Resume Protocol
Resume SHALL be integrated with gateway/session handling while remaining opt-in. The confirmed-resume check SHALL execute independently of resume-intent detection, and successful resume SHALL immediately return to the caller without falling through to the normal agent handler.

#### Scenario: Confirmed resume executes independently of intent detection
- **WHEN** a request includes `confirmResume: true` and `resumeRunId` is non-empty
- **THEN** the system invokes `ResumeManager.Resume` immediately
- **AND** `DetectResumeIntent` is NOT called for this request
- **AND** a `return` ends the handler after broadcasting `agent.resume_confirmed`

#### Scenario: Resume uses shutdown-derived bounded context
- **WHEN** the resume manager executes `FindCandidates` or `Resume`
- **THEN** the context is derived from the server's shutdown context and bounded by request timeout / hard ceiling configuration
- **AND** the operation respects server shutdown cancellation signals

#### Scenario: Resume stale TTL from config
- **WHEN** `NewResumeManager` is created
- **THEN** its stale TTL uses `config.RunLedger.StaleTTL`
- **AND** hardcoded `time.Hour` is not used

### Requirement: Tool Governance
The system SHALL expose tools to execution agents according to the active step's `ToolProfile`. The `run_*` tool namespace SHALL NOT be blanket-allowed; instead, run tools SHALL follow per-role allowlists consistent with the Access Control requirement.

#### Scenario: Coding profile tool access
- **WHEN** the active step uses the `coding` profile
- **THEN** only exact-match coding tools are available (e.g., `exec`, `exec_bg`, `exec_status`, `exec_stop`, `fs_read`, `fs_list`, `fs_write`, `fs_edit`, `fs_mkdir`, `fs_delete`, `fs_stat`)
- **AND** `strings.HasPrefix` matching is NOT used for tool name resolution

#### Scenario: Supervisor profile tool access
- **WHEN** the active step uses the `supervisor` profile
- **THEN** only `run_read`, `run_active`, and `run_note` are available

#### Scenario: Execution agent blocked from orchestrator run tools via profile guard
- **WHEN** a tool profile guard evaluates `run_create`, `run_apply_policy`, `run_approve_step`, or `run_resume` for an execution agent
- **THEN** access is denied
- **AND** only `run_read`, `run_active`, `run_note`, and `run_propose_step_result` pass the profile guard for execution agents

#### Scenario: Prefix matching does not grant unrelated tool access
- **WHEN** a tool named `execute_payment` exists and the active step uses the `coding` profile
- **THEN** `execute_payment` is NOT allowed
- **AND** only tools in the explicit coding tool set are allowed

### Requirement: Access Control
Tool access SHALL be role-based. The orchestrator (agent name `"orchestrator"` or `"lango-orchestrator"`) MAY call `run_create`, `run_apply_policy`, `run_approve_step`, `run_resume`. Execution agents MAY call `run_propose_step_result`. All agents MAY call `run_read`, `run_active`, `run_note`. An empty agent name SHALL NOT be treated as orchestrator identity.

#### Scenario: Empty agent name rejected
- **WHEN** a caller has an empty agent name (`""`)
- **THEN** `checkRole` returns `ErrAccessDenied` for orchestrator-only tools
- **AND** the caller is NOT silently granted orchestrator privilege

#### Scenario: System caller requires explicit identity
- **WHEN** an internal system component needs orchestrator-level access
- **THEN** it MUST set `SystemCallerName` (or an equivalent explicit identity) in the context
- **AND** the empty-string identity is never treated as a valid caller

#### Scenario: Execution agent blocked from run_create
- **WHEN** a non-orchestrator agent calls `run_create`
- **THEN** `ErrAccessDenied` is returned

### Requirement: Command Context
The system SHALL inject only active run summaries into command context. Completed, failed, and stale runs SHALL NOT be included in LLM context injection.

#### Scenario: Only active/paused runs injected
- **WHEN** the run summary provider assembles context for LLM injection
- **THEN** only runs with status `running` or `paused` are included
- **AND** runs with status `completed`, `failed`, or `stale` are excluded

#### Scenario: Section header matches content
- **WHEN** the assembled section is titled "Active Runs"
- **THEN** all listed runs have non-terminal status

### Requirement: Configuration
The system SHALL provide `RunLedgerConfig` under the root config with fields: `enabled`, `shadow`, `writeThrough`, `authoritativeRead`, `workspaceIsolation`, `staleTtl` (default: 1h), `validatorTimeout` (default: 2m), `plannerMaxRetries` (default: 2), `maxRunHistory`. All config values SHALL be wired to their respective runtime consumers.

#### Scenario: ValidatorTimeout applied as context deadline
- **WHEN** `PEVEngine.Verify` is invoked
- **THEN** the validator execution uses a context with a deadline set to `config.RunLedger.ValidatorTimeout`
- **AND** validators that exceed the timeout return a context deadline exceeded error

#### Scenario: MaxRunHistory triggers store-level pruning
- **WHEN** a run transitions to a terminal status (`completed` or `failed`)
- **THEN** the system invokes `PruneOldRuns(ctx, config.RunLedger.MaxRunHistory)`
- **AND** runs exceeding the history limit are removed from the store (oldest first)

## ADDED Requirements

### Requirement: Session Key Structured Context
The system SHALL provide a structured `RunContext` type in a shared package for workflow/background session identification, replacing fragile string-split parsing of session keys.

#### Scenario: Workflow session provides RunContext
- **WHEN** a workflow session is created with session key `"workflow:wf-123:run-456"`
- **THEN** a `RunContext` struct with `SessionType="workflow"`, `WorkflowID="wf-123"`, `RunID="run-456"` is stored in the request context

#### Scenario: Background session provides RunContext
- **WHEN** a background session is created with session key `"bg:run-789"`
- **THEN** a `RunContext` struct with `SessionType="background"`, `RunID="run-789"` is stored in the request context

#### Scenario: Guard reads RunContext instead of parsing session key
- **WHEN** `runIDFromSessionContext` resolves the run ID
- **THEN** it reads from the `RunContext` context value
- **AND** colon-split string parsing is NOT used

### Requirement: Store-Level Run Pruning
The system SHALL support pruning old runs from the store to enforce `MaxRunHistory`.

#### Scenario: PruneOldRuns removes excess runs
- **GIVEN** the store contains 150 runs and `MaxRunHistory` is 100
- **WHEN** `PruneOldRuns(ctx, 100)` is invoked
- **THEN** the 50 oldest completed/failed runs are removed
- **AND** active/paused runs are never pruned regardless of age

#### Scenario: PruneOldRuns is a no-op when under limit
- **GIVEN** the store contains 50 runs and `MaxRunHistory` is 100
- **WHEN** `PruneOldRuns(ctx, 100)` is invoked
- **THEN** no runs are removed
