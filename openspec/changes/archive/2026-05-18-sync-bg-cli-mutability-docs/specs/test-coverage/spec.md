## MODIFIED Requirements

### Requirement: Background CLI boundary guards stay executable
Repository-level regressions in background CLI boundary messaging SHALL be enforced by executable tests.

#### Scenario: Root bg boundary and docs guards reject misleading references
- **WHEN** root CLI bg commands are wired without an in-process manager
- **THEN** executable tests SHALL fail if the error implies `lango serve` alone makes standalone `lango bg` work
- **AND** executable docs guards SHALL fail if public docs list `lango bg` commands without the in-memory/root-CLI boundary caveat

#### Scenario: Background automation docs guard rejects read-only CLI wording
- **WHEN** docs quality tests run
- **THEN** `docs/automation/background.md` SHALL fail the test suite if it describes all `lango bg` CLI commands as read-only
- **AND** it SHALL be checked for wording that distinguishes inspect-only commands from `lango bg cancel <id>` requesting cancellation in the same process
