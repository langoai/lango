## ADDED Requirements

### Requirement: Public workbench docs use neutral completed-turn wording

Public workbench documentation SHALL describe the completed-turn state with neutral wording that does not imply success.

#### Scenario: Docs say finished rather than complete
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL describe the previous turn as finished rather than completed
