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
