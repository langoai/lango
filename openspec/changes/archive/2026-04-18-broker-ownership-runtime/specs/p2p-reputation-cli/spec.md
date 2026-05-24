## ADDED Requirements

### Requirement: Reputation CLI supports broker-backed runtime reads
The `lango p2p reputation` command MUST remain functional when runtime reputation reads are served through broker-backed storage.

#### Scenario: Reputation CLI under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** `lango p2p reputation` returns reputation details through broker-backed storage capabilities
