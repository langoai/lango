# background-hygiene-compaction Specification

## Purpose
TBD - created by archiving change ux-continuity. Update Purpose after archive.
## Requirements
### Requirement: CompactionBuffer async buffer
The system SHALL provide a `session.CompactionBuffer` in `internal/session/compaction_buffer.go` that accepts compaction jobs via `EnqueueCompaction(key string, upToIndex int)` and executes them asynchronously on a bounded worker pool. The buffer SHALL follow the same lifecycle contract as `learning.AnalysisBuffer`: bounded queue, drop-with-warning on overflow, graceful `Drain(timeout time.Duration) error` on shutdown, and `Start(ctx)` / `Stop()` hooks integrated with `lifecycle.Registry`.

#### Scenario: Enqueue runs compaction asynchronously
- **WHEN** `EnqueueCompaction("sess-1", 20)` is called
- **THEN** the call SHALL return immediately without blocking
- **AND** a background worker SHALL invoke `EntStore.CompactMessages("sess-1", 20, <summary>)` within the worker budget

#### Scenario: Overflow drops with warning
- **WHEN** the queue is full and `EnqueueCompaction` is called again
- **THEN** the job SHALL be dropped
- **AND** a warning SHALL be logged including the session key and current queue depth

#### Scenario: Drain completes before shutdown
- **WHEN** `Drain(3 * time.Second)` is called with pending jobs
- **THEN** the buffer SHALL wait up to 3 seconds for in-flight jobs to finish
- **AND** SHALL return `nil` if all jobs complete
- **AND** SHALL return a timeout error otherwise, leaving remaining jobs cancelled

### Requirement: Post-turn compaction trigger
After every `TurnCompletedEvent`, the system SHALL estimate the total tokens of the session's current messages using `types.EstimateTokens()`. When the estimate exceeds `modelWindow * context.compaction.threshold` (default 0.5), the system SHALL enqueue a compaction job via `CompactionBuffer.EnqueueCompaction`.

#### Scenario: Threshold exceeded enqueues compaction
- **WHEN** a turn completes with total estimated tokens at 55% of the model window
- **AND** `context.compaction.enabled` is `true` (default)
- **THEN** a compaction job SHALL be enqueued for that session

#### Scenario: Below threshold does not enqueue
- **WHEN** a turn completes with total estimated tokens at 30% of the model window
- **THEN** no compaction job SHALL be enqueued

#### Scenario: Feature disabled does not enqueue
- **WHEN** `context.compaction.enabled` is `false`
- **THEN** no compaction job SHALL be enqueued regardless of token usage

### Requirement: Sync-point guard at turn start
`ContextAwareModelAdapter.GenerateContent()` SHALL check for an in-flight compaction on the current session key before assembling context. It SHALL wait up to `context.compaction.syncTimeout` (default 2s) for that compaction to complete. On timeout the method SHALL log a warning, emit `CompactionSlowEvent`, and proceed with the current session state.

#### Scenario: In-flight compaction completes within timeout
- **WHEN** a new turn starts while a compaction for the same session is running
- **AND** the compaction finishes within 2s
- **THEN** `GenerateContent()` SHALL observe the compacted message list before building sections

#### Scenario: In-flight compaction exceeds timeout
- **WHEN** an in-flight compaction has not completed within 2s
- **THEN** `GenerateContent()` SHALL proceed with the current messages
- **AND** SHALL log a warning with session key
- **AND** SHALL emit a `CompactionSlowEvent` on the eventbus

#### Scenario: No in-flight compaction skips wait entirely
- **WHEN** no compaction is in flight for the current session
- **THEN** `GenerateContent()` SHALL NOT wait and SHALL proceed without delay

### Requirement: Compaction does not trigger on budgets.Degraded
The post-turn compaction trigger SHALL NOT be activated solely because `budgets.Degraded` is `true`. The `Degraded` flag indicates that the base prompt is too large for the model window (a configuration issue) and session message compaction cannot resolve it.

#### Scenario: Degraded without excess tokens is a no-op for compaction
- **WHEN** a turn completes with `budgets.Degraded == true` but session message tokens below the threshold
- **THEN** no compaction job SHALL be enqueued
- **AND** the existing Degraded warning log from `context-budget` spec SHALL continue to be emitted by the context model

### Requirement: CompactionCompletedEvent
On successful completion of a compaction job, the buffer SHALL publish a `CompactionCompletedEvent` on the eventbus with fields `SessionKey string`, `UpToIndex int`, `SummaryTokens int`, `ReclaimedTokens int`, and `Timestamp time.Time`.

#### Scenario: Event published on success
- **WHEN** a compaction job successfully replaces messages 0..20 with a summary
- **THEN** a `CompactionCompletedEvent` SHALL be published
- **AND** the event SHALL contain the session key, `UpToIndex=20`, summary token count, and reclaimed token count

#### Scenario: Event not published on failure
- **WHEN** a compaction job fails (database error, empty summary)
- **THEN** no `CompactionCompletedEvent` SHALL be published
- **AND** a warning SHALL be logged with the failure reason

### Requirement: Config surface for compaction
The system SHALL provide additive fields under `context.compaction`: `enabled bool` (default `true`), `threshold float64` (default `0.5`, valid range `[0.1, 0.95]`), `syncTimeout time.Duration` (default `2s`, valid range `[100ms, 10s]`), and `workerCount int` (default `1`). Invalid values SHALL be clamped to the valid range with a warning log.

#### Scenario: Default configuration
- **WHEN** no `context.compaction.*` config is set
- **THEN** compaction SHALL be enabled with threshold 0.5, sync timeout 2s, and 1 worker

#### Scenario: Invalid threshold clamped
- **WHEN** `context.compaction.threshold` is set to `1.5`
- **THEN** the effective threshold SHALL be clamped to `0.95` with a warning log

