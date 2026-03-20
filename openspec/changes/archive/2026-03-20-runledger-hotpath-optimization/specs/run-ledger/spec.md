## MODIFIED Requirements

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

## ADDED Requirements

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
