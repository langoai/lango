## ADDED Requirements

### Requirement: Public workbench docs describe the completed-turn body as a next-step state

Public workbench documentation SHALL explain that the completed-turn empty body shifts from a blank no-missions message to a next-step prompt.

#### Scenario: Docs mention completed-turn body shift
- **WHEN** a user reads the README or CLI/TUI docs for the standalone workbench
- **THEN** those docs SHALL mention that a completed turn changes the empty body into a next-step prompt
