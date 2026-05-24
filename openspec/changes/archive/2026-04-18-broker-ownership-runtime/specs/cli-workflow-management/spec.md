## ADDED Requirements

### Requirement: Workflow history/status support broker-backed runtime reads
Workflow CLI read surfaces MUST remain functional when runtime state is served by broker-backed storage.

#### Scenario: Workflow read path under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** workflow list/history/status read state through broker-backed storage capabilities
