# RunLedger — Task OS Durable Execution Engine

## Purpose

Durable execution engine that transforms Lango from an AI chatbot into a Task OS. Provides an append-only journal as the single source of truth, typed validators via a Propose-Evidence-Verify (PEV) engine, and policy-driven failure recovery for long-running agent tasks. Core principle: "the system proves completion, not the agent."
## Requirements
### Requirement: Append-Only Journal
The system SHALL provide an Ent-backed persistent journal implementation for RunLedger. `AppendJournalEvent` SHALL NOT acquire a Go-level mutex — the database transaction and `(run_id, seq)` unique constraint SHALL provide serialization.

#### Scenario: Journal survives restart
- **WHEN** the process restarts after journal events are appended
- **THEN** the same run's journal events remain queryable

#### Scenario: App runtime prefers Ent store
- **WHEN** the shared application `ent.Client` is available during RunLedger module init
- **THEN** the module uses an Ent-backed `RunLedgerStore`
- **AND** `MemoryStore` remains a fallback for tests and non-bootstrapped contexts

#### Scenario: Concurrent journal appends to same run
- **WHEN** two goroutines concurrently call `AppendJournalEvent` for the same run
- **THEN** both events SHALL be persisted with distinct sequence numbers
- **AND** the `(run_id, seq)` unique constraint SHALL prevent duplicate sequences

#### Scenario: Concurrent journal appends to different runs
- **WHEN** two goroutines concurrently call `AppendJournalEvent` for different runs
- **THEN** neither goroutine SHALL block the other at the Go-level
- **AND** both events SHALL be persisted independently

### Requirement: Materialized Snapshots
The system SHALL support authoritative-read mode where run-state reads come from RunLedger snapshots. The in-memory snapshot cache SHALL use per-run locking instead of a global mutex, so that operations on different runs do not contend.

#### Scenario: Authoritative snapshot read
- **WHEN** authoritative-read is enabled
- **THEN** run-state consumers read from `RunSnapshot`
- **AND** projection mirrors are no longer treated as authoritative

#### Scenario: Per-run cache isolation
- **WHEN** two goroutines concurrently call `GetCachedSnapshot` for different run IDs
- **THEN** neither goroutine SHALL block the other
- **AND** each goroutine SHALL receive the correct snapshot for its run

#### Scenario: Per-run cache write safety
- **WHEN** two goroutines concurrently call `UpdateCachedSnapshot` and `GetCachedSnapshot` for the same run ID
- **THEN** the per-run lock SHALL serialize access
- **AND** no data race SHALL occur

### Requirement: Run Lifecycle
`checkRunCompletion` SHALL only journal `EventCriterionMet` for criteria that are **newly** met — criteria that transitioned from `Met=false` to `Met=true` during the current verification pass.

#### Scenario: Already-met criteria are not re-journaled
- **GIVEN** a run where criterion 0 was already met (journaled in a previous pass)
- **WHEN** `checkRunCompletion` runs and criterion 1 is newly met
- **THEN** only one `EventCriterionMet` journal entry is appended (for criterion 1)
- **AND** no duplicate entry is created for criterion 0

#### Scenario: First-time met criteria are journaled
- **GIVEN** a run where no criteria have been met yet
- **WHEN** `checkRunCompletion` runs and criteria 0 and 2 pass validation
- **THEN** exactly two `EventCriterionMet` journal entries are appended (for indices 0 and 2)

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
`VerifyAcceptanceCriteria` SHALL NOT mutate its input slice. It SHALL return both the list of unmet criteria and a fully evaluated copy of the criteria slice.

#### Scenario: Input criteria slice is not mutated
- **GIVEN** a criteria slice where all items have `Met = false`
- **WHEN** `VerifyAcceptanceCriteria` is called and some criteria pass validation
- **THEN** the original criteria slice items still have `Met = false`
- **AND** the returned evaluated copy has `Met = true` for passing criteria

#### Scenario: MetAt is set on newly met criteria
- **WHEN** a criterion passes validation in `VerifyAcceptanceCriteria`
- **THEN** the evaluated copy has `Met = true` and `MetAt` set to the current time via `time.Now()`
- **AND** `MetAt` is not nil

#### Scenario: Dead ctxKeyNow code is removed
- **WHEN** `VerifyAcceptanceCriteria` sets `MetAt` on a passing criterion
- **THEN** it uses `time.Now()` directly
- **AND** no context-value lookup for time injection exists

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

### Requirement: Workspace Isolation
Production runtime SHALL activate workspace isolation for coding steps once the execution-isolation stage is enabled. Error messages SHALL provide guided remediation with actionable commands. Patch application failures SHALL include rollback instructions.

#### Scenario: Runtime isolation active
- **WHEN** `runLedger.workspaceIsolation` is enabled
- **THEN** the app runtime wires `PEVEngine.WithWorkspace(...)`
- **AND** coding-step validators execute inside isolated worktrees rather than the base tree

#### Scenario: Retry-safe repeated validation
- **WHEN** the same step is validated multiple times under isolation
- **THEN** each attempt uses a retry-safe workspace identity
- **AND** previous attempts do not block later ones via reused branch metadata

#### Scenario: Dirty tree guided remediation
- **WHEN** `CheckDirtyTree` detects uncommitted changes
- **THEN** the error message includes a count of changed files
- **AND** the error suggests `git stash push -m "lango-workspace-isolation"` as a remediation command

#### Scenario: Patch apply conflict guidance
- **WHEN** `ApplyPatch` fails due to a merge conflict
- **THEN** the error message includes the raw git output
- **AND** the error instructs the user to run `git am --abort` to rollback

#### Scenario: Enablement conditions
- **WHEN** the system evaluates whether workspace isolation should be active
- **THEN** isolation is required for steps with validators of type `file_changed`, `build_pass`, or `test_pass`
- **AND** isolation is not required for validators of type `human_approval` or `always_pass`

### Requirement: RunLedger Workspace Isolation doctor check
The doctor command SHALL include a `RunLedger Workspace Isolation` check that validates the workspace isolation configuration and environment health. The check name SHALL be distinct from the existing `P2P Workspaces` check.

#### Scenario: Isolation enabled and healthy
- **WHEN** `runLedger.workspaceIsolation` is enabled and git is available and no stale worktrees exist
- **THEN** the check status is `Pass`
- **AND** the message includes the config value and active worktree count

#### Scenario: Isolation disabled
- **WHEN** `runLedger.workspaceIsolation` is disabled
- **THEN** the check status is `Skip`
- **AND** the message indicates isolation is not enabled

#### Scenario: Git unavailable
- **WHEN** `runLedger.workspaceIsolation` is enabled but `git` is not in PATH
- **THEN** the check status is `Warn`
- **AND** the message indicates git is required for workspace isolation

#### Scenario: Stale worktrees detected
- **WHEN** `git worktree list` reports worktrees under the runledger temp directory that no longer exist on disk
- **THEN** the check status is `Warn`
- **AND** the message lists the stale worktree paths

#### Scenario: Doctor help text
- **WHEN** user runs `lango doctor --help`
- **THEN** the output lists `RunLedger Workspace Isolation` under the Execution category
- **AND** the total check count is incremented by 1

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
The system SHALL expose tools to execution agents according to the active step's `ToolProfile`. The system SHALL cache the `RunSnapshot` within a single turn's context to avoid redundant store lookups. First tool call in a turn SHALL fetch and cache; subsequent calls in the same turn SHALL reuse the cached snapshot.

#### Scenario: Coding profile
- **WHEN** the active step uses the `coding` profile
- **THEN** only coding-safe execution tools are available

#### Scenario: Supervisor profile
- **WHEN** the active step uses the `supervisor` profile
- **THEN** only supervisor-safe run inspection/approval tools are available

#### Scenario: Per-turn snapshot cache hit
- **WHEN** `ToolProfileGuard` is invoked for the 2nd through Nth tool call in the same turn
- **AND** a cached snapshot exists in the turn context for the same run ID
- **THEN** the guard SHALL use the cached snapshot without calling `store.GetRunSnapshot()`

#### Scenario: Per-turn snapshot cache miss
- **WHEN** `ToolProfileGuard` is invoked for the first tool call in a turn
- **AND** no cached snapshot exists in the turn context
- **THEN** the guard SHALL call `store.GetRunSnapshot()`, cache the result in the turn context, and proceed with profile checking

#### Scenario: Cache isolation across turns
- **WHEN** a new turn begins with a fresh context
- **THEN** the snapshot cache from the previous turn SHALL NOT be reused
- **AND** the first tool call in the new turn SHALL fetch a fresh snapshot

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

### Requirement: Ent Schemas
The system SHALL provide 3 Ent schemas: `RunJournal` (append-only event log with run_id+seq unique index), `RunSnapshot` (cached materialized view with unique run_id), `RunStep` (step projection with run_id+step_id unique index).

#### Scenario: Journal uniqueness
- **GIVEN** the RunJournal schema
- **WHEN** two events with the same run_id and seq are inserted
- **THEN** a unique constraint violation occurs

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

### Requirement: CLI Journal Inspection
The system SHALL let operators inspect persistent RunLedger data from the CLI.

#### Scenario: `lango run list`
- **WHEN** the operator runs `lango run list`
- **THEN** the command reads recent runs from the persistent RunLedger snapshot store

#### Scenario: `lango run journal <run-id>`
- **WHEN** the operator runs `lango run journal <run-id>`
- **THEN** the command reads the persistent journal events for that run

### Requirement: Command Context
The system SHALL inject active run summaries into command context. The system SHALL cache assembled run summary strings per session with journal-sequence-based invalidation to avoid redundant queries on repeated LLM requests.

#### Scenario: Active run summary injected
- **WHEN** an active or paused resumable run exists for the session
- **THEN** command context includes compact run summary, current blocker, and current step data

#### Scenario: Summary cache hit
- **WHEN** `assembleRunSummarySection` is called for a session
- **AND** a cached summary exists for that session
- **AND** the maximum journal sequence for the session's runs has not changed since the cache was populated
- **THEN** the system SHALL return the cached summary string without querying the run summary store

#### Scenario: Summary cache invalidation on journal change
- **WHEN** `assembleRunSummarySection` is called for a session
- **AND** a cached summary exists for that session
- **AND** the maximum journal sequence for the session's runs has increased since the cache was populated
- **THEN** the system SHALL discard the cached entry, query fresh summaries, assemble the string, and update the cache

#### Scenario: Summary cache miss
- **WHEN** `assembleRunSummarySection` is called for a session
- **AND** no cached summary exists for that session
- **THEN** the system SHALL query summaries from the store, assemble the string, store it in the cache with the current max journal sequence, and return it

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

### Requirement: Snapshot Deep Copy
The system SHALL provide a `RunSnapshot.DeepCopy()` method that returns a fully independent copy of the snapshot with no shared mutable state.

#### Scenario: DeepCopy produces independent snapshot
- **WHEN** `DeepCopy()` is called on a snapshot with steps, acceptance criteria, and notes
- **THEN** the returned snapshot has the same field values
- **AND** appending to `copy.Steps` does not affect the original's `Steps` slice
- **AND** modifying `copy.Notes["key"]` does not affect the original's `Notes` map

#### Scenario: DeepCopy preserves MetAt pointer semantics
- **GIVEN** a snapshot with an `AcceptanceCriterion` where `MetAt` points to a time value
- **WHEN** `DeepCopy()` is called
- **THEN** the copy's `MetAt` is a new pointer with the same time value
- **AND** modifying the copy's `MetAt` does not affect the original

#### Scenario: DeepCopy produces independent SourceDescriptor
- **WHEN** `DeepCopy()` is called on a snapshot with a non-nil `SourceDescriptor`
- **THEN** modifying the copy's `SourceDescriptor` backing array SHALL NOT affect the original

#### Scenario: DeepCopy resets step index
- **WHEN** `DeepCopy()` is called on a snapshot with a warm step index
- **THEN** the copy's step index SHALL be nil (lazy rebuild on next FindStep)

### Requirement: Lazy step index for FindStep
`RunSnapshot.FindStep(stepID)` SHALL use a lazily-built `map[string]int` index for O(1) lookup. The index SHALL be invalidated (set to nil) when `Steps` is mutated by `EventPlanAttached` or `PolicyDecompose`. The index SHALL NOT be serialized to JSON (`json:"-"` tag). After JSON unmarshal or `DeepCopy()`, the index SHALL be nil and rebuilt on next `FindStep` call.

#### Scenario: FindStep after PlanAttached
- **WHEN** a snapshot is materialized from journal events including `plan_attached`
- **THEN** `FindStep` SHALL return the correct step by ID in O(1) time

#### Scenario: FindStep after PolicyDecompose adds new steps
- **WHEN** `PolicyDecompose` appends new steps to the snapshot
- **THEN** `FindStep` SHALL find both original and newly-added steps

#### Scenario: FindStep after DeepCopy
- **WHEN** `DeepCopy()` is called and the original's steps are mutated
- **THEN** the copy's `FindStep` SHALL return its own independent step data

#### Scenario: FindStep after JSON round-trip
- **WHEN** a snapshot is marshaled to JSON and unmarshaled back
- **THEN** `FindStep` SHALL work correctly via lazy index rebuild

### Requirement: SourceKind and SourceDescriptor in RunCreatedPayload
`RunCreatedPayload` SHALL include `SourceKind string` (values: "workflow", "background", "") and `SourceDescriptor json.RawMessage` (original workflow or origin JSON). These fields SHALL be persisted in the journal and restored into `RunSnapshot` during event replay.

#### Scenario: Workflow run stores source descriptor
- **WHEN** `WorkflowWriteThrough.CreateRun()` creates a run
- **THEN** the journal event SHALL include `SourceKind: "workflow"` and the workflow marshaled as `SourceDescriptor`

#### Scenario: Background task stores source descriptor
- **WHEN** `BackgroundWriteThrough.PrepareTask()` creates a run
- **THEN** the journal event SHALL include `SourceKind: "background"` and the origin marshaled as `SourceDescriptor`

#### Scenario: Legacy journals without SourceKind
- **WHEN** a journal event from before this change is replayed
- **THEN** `SourceKind` SHALL be empty string and `SourceDescriptor` SHALL be nil (zero values)

### Requirement: Marshal Payload Error Observability
`marshalPayload` SHALL log a warning when JSON marshaling fails instead of silently returning an empty object.

#### Scenario: Marshal failure is logged
- **WHEN** `marshalPayload` receives a value that fails `json.Marshal` (e.g., channel type, cyclic reference)
- **THEN** a warning-level log message is emitted containing the error details
- **AND** the function still returns `{}` as a fallback
- **AND** the caller's flow is not interrupted

#### Scenario: Successful marshal is not logged
- **WHEN** `marshalPayload` receives a valid serializable value
- **THEN** no log message is emitted
- **AND** the correct JSON bytes are returned

### Requirement: Projection Sync Error Observability
Projection sync errors in `writethrough.go` SHALL be logged at warning level instead of being discarded with `_ =`.

#### Scenario: Degraded projection sync error is logged
- **WHEN** `appendProjectionSyncEvent` returns an error in a write-through method
- **THEN** a warning-level log message is emitted containing the run ID and error details
- **AND** the outer operation continues (best-effort semantics preserved)

#### Scenario: Successful projection sync is not logged
- **WHEN** `appendProjectionSyncEvent` returns nil
- **THEN** no warning log message is emitted

### Requirement: Snapshot Lookup Benchmarks
The system SHALL include benchmarks that measure ToolProfileGuard performance with and without the per-turn context-scoped cache.

#### Scenario: Benchmark cached vs uncached guard
- **WHEN** the benchmark suite runs `BenchmarkToolProfileGuard_WithCache` and `BenchmarkToolProfileGuard_NoCache`
- **THEN** the cached variant SHALL complete in fewer allocations and less time per operation than the uncached variant for N > 1 tool calls per turn

### Requirement: Summary Assembly Benchmarks
The system SHALL include benchmarks that measure `assembleRunSummarySection` performance with and without the session-scoped cache.

#### Scenario: Benchmark cache hit vs cache miss
- **WHEN** the benchmark suite runs `BenchmarkAssembleRunSummary_CacheHit` and `BenchmarkAssembleRunSummary_CacheMiss`
- **THEN** the cache-hit variant SHALL demonstrate reduced allocations and query count compared to the cache-miss variant

### Requirement: EntStore Concurrency Benchmarks
The system SHALL include benchmarks that measure EntStore performance under parallel run access with per-run locks versus the baseline global mutex.

#### Scenario: Benchmark parallel runs
- **WHEN** the benchmark suite runs `BenchmarkEntStore_ParallelRuns` with M concurrent runs and N goroutines per run
- **THEN** the per-run lock variant SHALL demonstrate reduced total elapsed time compared to the global mutex baseline when M > 1

### Requirement: Max Journal Sequence Query
The store SHALL provide a method to retrieve the maximum journal sequence number for a session's runs, for use in cache invalidation.

#### Scenario: Max sequence for active session
- **WHEN** `MaxJournalSeqForSession` is called with a session key that has active runs
- **THEN** the method SHALL return the highest `last_journal_seq` value across all run snapshots for that session

#### Scenario: Max sequence for empty session
- **WHEN** `MaxJournalSeqForSession` is called with a session key that has no runs
- **THEN** the method SHALL return 0 and no error

### Requirement: Store Option Pattern
MemoryStore and EntStore constructors SHALL accept variadic `StoreOption` parameters via `WithAppendHook(func(JournalEvent))`. The `RunLedgerStore` interface SHALL NOT be modified.

#### Scenario: Backward compatible construction
- **WHEN** `NewMemoryStore()` or `NewEntStore(client)` is called without options
- **THEN** behavior is identical to pre-change behavior

#### Scenario: Append hook registration
- **WHEN** `NewMemoryStore(WithAppendHook(h))` is called
- **THEN** the hook `h` is called after each successful journal event append

#### Scenario: Hook runs outside lock
- **WHEN** an append hook reads from the same MemoryStore it is registered on
- **THEN** no deadlock occurs because the hook is invoked after the write lock is released

### Requirement: AppendHookSetter Interface
Concrete store types (`MemoryStore`, `EntStore`) SHALL implement the `AppendHookSetter` interface with a `SetAppendHook(func(JournalEvent))` method for post-construction hook registration. This interface is NOT part of the `RunLedgerStore` contract.

#### Scenario: Post-construction hook registration
- **WHEN** `SetAppendHook` is called on a store after construction
- **THEN** the registered hook is invoked on subsequent `AppendJournalEvent` calls

#### Scenario: Hook chaining preserves existing hooks
- **WHEN** a store is created with `WithAppendHook(first)` and then `SetAppendHook(second)` is called
- **THEN** both `first` and `second` are invoked in order on each `AppendJournalEvent` call

### Requirement: Runtime wake boundary definition
The system SHALL document the boundary between application-layer resume (opt-in `confirmResume + resumeRunId` handshake) and runtime-layer wake (harness re-initialization from persisted state). The design document SHALL enumerate the state categories that must persist for wake to be possible without a full bootstrap pipeline.

#### Scenario: State categories enumerated
- **WHEN** the design document is reviewed
- **THEN** it explicitly covers: in-flight tool call state, pending approval state, supervisor/ADK session bridge state, and crypto provider re-initialization
- **AND** it maps which of these the current resume protocol covers vs does not cover

#### Scenario: No runtime behavior change
- **WHEN** this change is implemented
- **THEN** no runtime code paths for session handling, resume, or bootstrap are modified
- **AND** the change is limited to design documentation and diagnostic tooling
