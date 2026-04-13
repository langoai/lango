## MODIFIED Requirements

### Requirement: Unified bootstrap sequence

#### Scenario: Load security state before envelope migration
- **WHEN** the bootstrap sequence runs
- **THEN** `phaseLoadSecurityState` SHALL execute before `phaseMigrateEnvelope`
- **AND** when an envelope has `PendingMigration` or `PendingRekey`, salt and checksum SHALL be loaded even when envelope is present
