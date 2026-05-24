## ADDED Requirements

### Requirement: Background CLI boundary guards stay executable
Repository-level regressions in background CLI boundary messaging SHALL be enforced by executable tests.

#### Scenario: Root bg boundary and docs guards reject misleading references
- **WHEN** root CLI bg commands are wired without an in-process manager
- **THEN** executable tests SHALL fail if the error implies `lango serve` alone makes standalone `lango bg` work
- **AND** executable docs guards SHALL fail if public docs list `lango bg` commands without the in-memory/root-CLI boundary caveat
