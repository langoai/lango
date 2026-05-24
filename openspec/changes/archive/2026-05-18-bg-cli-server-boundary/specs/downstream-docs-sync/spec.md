## ADDED Requirements

### Requirement: Background CLI docs describe server-boundary caveat
Public documentation that lists `lango bg` commands SHALL explain that the current root CLI is not a remote client for the server process's in-memory background manager.

#### Scenario: Public bg command references include runtime caveat
- **WHEN** a user reads README, `docs/cli/index.md`, or `docs/automation/background.md`
- **THEN** any `lango bg list/status/cancel/result` command reference SHALL be accompanied by a caveat that task state is in-memory and root CLI management is not yet connected to the running server process
