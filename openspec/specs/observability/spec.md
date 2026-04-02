## ADDED Requirements

### Requirement: Audit recorder handles AlertEvent
The audit recorder SHALL subscribe to AlertEvent via SubscribeTyped and persist each alert to the audit log with action="alert", actor="system", target=alert type, and details containing severity, message, and alert-specific metadata.

#### Scenario: AlertEvent persisted to audit log
- **WHEN** an AlertEvent is published on the EventBus
- **THEN** the audit recorder creates an audit log entry with action="alert", actor="system", target set to the alert type, and details containing severity and message

### Requirement: Alerts HTTP route registered
The `/alerts` HTTP route SHALL be registered alongside existing observability routes when the observability system is enabled and a database client is available.

#### Scenario: Alerts endpoint available
- **WHEN** observability is enabled and the application starts
- **THEN** the GET `/alerts` endpoint is registered on the chi router


## ADDED Requirements

### Requirement: Session map capacity limit
The `MetricsCollector` MUST support a `MaxSessions` field (default: 10,000) that caps the number of tracked sessions. When the cap is reached and a new session is inserted, the least-recently-updated session MUST be evicted.

#### Scenario: Eviction at capacity
- **WHEN** `MaxSessions` is 10,000 and the 10,001st session records token usage
- **THEN** the oldest session (by `LastUpdated`) is removed before the new session is inserted

#### Scenario: Eviction selects oldest
- **GIVEN** sessions A (updated 1 min ago) and B (updated 5 min ago) at capacity
- **WHEN** a new session C records usage
- **THEN** session B is evicted (oldest `LastUpdated`)

#### Scenario: Cap disabled
- **WHEN** `MaxSessions` is 0 or negative
- **THEN** no eviction occurs and sessions grow unbounded (backward compatible)

### Requirement: Session metric timestamp
Each `SessionMetric` MUST include a `LastUpdated time.Time` field that is set to `time.Now()` on every `RecordTokenUsage` call for that session.
