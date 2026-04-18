## ADDED Requirements

### Requirement: Alert history supports broker-backed runtime reads
The observability alerts route MUST remain functional when runtime alert history is served by broker-backed storage.

#### Scenario: Alerts route under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** the `/alerts` route returns alert history through broker-backed storage capabilities
