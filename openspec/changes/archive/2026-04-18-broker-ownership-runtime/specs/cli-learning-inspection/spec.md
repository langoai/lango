## ADDED Requirements

### Requirement: Learning history supports broker-backed runtime reads
The `lango learning history` command MUST remain functional when bootstrap is broker-owned and runtime reads come from broker-backed storage.

#### Scenario: Learning history under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** `lango learning history` still returns recent learning records through the broker-backed reader path
