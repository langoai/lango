## Purpose

Delta spec for session-local budget persistence replacing process-shared budget serialization.

## Requirements

### Requirement: Budget persistence uses session-local state

Budget counters SHALL be tracked per-session in `budgetRestoringExecutor.sessionState`, not from the process-shared `BudgetPolicy` instance. The `CoordinatingExecutor` SHALL expose per-run budget stats keyed by session ID via `LastRunStatsForSession(sessionID)` with consume-once semantics.

#### Scenario: Cumulative turns accumulate across runs
- **WHEN** a session restores with baseline turns=7 and a run produces 3 turns
- **THEN** the persisted `usage:budget_turns` SHALL be 10

#### Scenario: Cross-session isolation
- **WHEN** two sessions with different keys run concurrently
- **THEN** each session's budget counters SHALL be independent
- **AND** one session's stats SHALL NOT overwrite the other's

#### Scenario: Concurrent stats storage
- **WHEN** `CoordinatingExecutor.RunStreamingDetailed` completes for session A
- **AND** session B is running concurrently
- **THEN** session A's stats SHALL be stored independently via `sync.Map[sessionID → RunStats]`

### Requirement: Policy event publishing is unconditional

The `policyBus` SHALL be initialized whenever the event bus exists (`bus != nil`), regardless of `cfg.Hooks.EventPublishing` or `cfg.Hooks.Enabled` settings. Policy observability SHALL work in default single-agent configuration.

#### Scenario: Default config publishes policy events
- **WHEN** the app runs with default config (single-agent, hooks defaults)
- **AND** observability is enabled
- **THEN** policy decision events SHALL be published to audit and metrics
