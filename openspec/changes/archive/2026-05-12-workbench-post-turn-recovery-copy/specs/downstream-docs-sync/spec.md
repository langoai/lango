## ADDED Requirements

### Requirement: Public workbench docs mention recovery wording after failed turns

Public workbench documentation SHALL explain that failed completed-turn states switch to recovery-specific wording.

#### Scenario: Docs mention recovery wording
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that failed turns pivot the next prompt loop into recovery wording
