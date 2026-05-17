## MODIFIED Requirements

### Requirement: Background CLI docs describe gateway-backed runtime boundary
Public documentation that lists `lango bg` commands SHALL explain that root CLI background management talks to the running Lango gateway and still manages only that process's in-memory task state.

#### Scenario: Public bg command references include gateway runtime caveat
- **WHEN** a user reads README, `docs/cli/index.md`, or `docs/automation/background.md`
- **THEN** any `lango bg list/status/cancel/result` command reference SHALL be accompanied by a caveat that task state is in-memory and lost on server restart
- **AND** the docs SHALL state that root CLI management requires a reachable Lango gateway
- **AND** the docs SHALL mention the `--addr` override for selecting the gateway address

#### Scenario: Background automation docs describe cancel mutability
- **WHEN** a user reads `docs/automation/background.md`
- **THEN** the document SHALL NOT describe all `lango bg` CLI commands as read-only management commands
- **AND** it SHALL distinguish inspect-only `list/status/result` behavior from `cancel` requesting cancellation for a pending or running task in the target gateway process
