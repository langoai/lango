## MODIFIED Requirements

### Requirement: Turn trace Store interface
The turn trace Store interface SHALL include the following additional methods beyond the existing `CreateTrace`, `AppendEvent`, `FinishTrace`, `RecentFailures`, and `IsolationLeakCount`:

- `EventsForTrace(ctx context.Context, traceID string) ([]Event, error)` — returns all events for a trace, ordered by seq
- `TracesForSession(ctx context.Context, sessionKey string) ([]Trace, error)` — returns all traces for a session, ordered by started_at
- `PurgeTraces(ctx context.Context, traceIDs []string) error` — deletes traces and their associated events
- `TraceCount(ctx context.Context) (int, error)` — returns total trace count
- `OldTraces(ctx context.Context, cutoff time.Time, onlySuccess bool, limit int) ([]string, error)` — returns trace IDs older than cutoff
- `RecentByOutcome(ctx context.Context, outcome Outcome, since time.Time, limit int) ([]Trace, error)` — returns traces matching outcome within time window

All methods SHALL be implemented in `EntStore`. All methods SHALL be nil-safe (return nil/0 when store is nil).

#### Scenario: Query events for trace
- **WHEN** `EventsForTrace` is called with a valid trace ID
- **THEN** it SHALL return all events ordered by sequence number

#### Scenario: Query traces by outcome and time window
- **WHEN** `RecentByOutcome` is called with `OutcomeLoopDetected` and `since` 24 hours ago
- **THEN** it SHALL return only traces with that outcome created after the cutoff

#### Scenario: Purge cascades to events
- **WHEN** `PurgeTraces` is called with trace IDs
- **THEN** both the trace rows and their associated event rows SHALL be deleted

#### Scenario: Nil store returns safely
- **WHEN** any method is called on a nil `EntStore`
- **THEN** it SHALL return nil/0/empty without error
