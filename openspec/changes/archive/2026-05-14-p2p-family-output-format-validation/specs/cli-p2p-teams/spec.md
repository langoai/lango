## ADDED Requirements
### Requirement: Team guidance JSON output uses explicit output selection
The team guidance CLI SHALL expose machine-readable output through `--output json` instead of a boolean `--json` toggle.

#### Scenario: Team list and status JSON output remain machine-readable
- **WHEN** the user runs `lango p2p team list --output json` or `lango p2p team status <team-id> --output json`
- **THEN** the commands SHALL emit the same machine-readable guidance payloads as before
