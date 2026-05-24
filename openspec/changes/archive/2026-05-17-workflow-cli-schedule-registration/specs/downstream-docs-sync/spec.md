## MODIFIED Requirements

### Requirement: Public workflow docs describe scheduled workflow registration truthfully
Public automation documentation SHALL describe the current behavior of `lango workflow run --schedule` without stale "not implemented" wording.

#### Scenario: CLI automation docs show cron-backed registration
- **WHEN** `docs/cli/automation.md` documents `lango workflow run --schedule`
- **THEN** it SHALL explain that the command validates the workflow and registers an enabled cron job that asks runtime automation to invoke `workflow_run`
- **AND** it SHALL NOT instruct operators that CLI schedule registration is unavailable
