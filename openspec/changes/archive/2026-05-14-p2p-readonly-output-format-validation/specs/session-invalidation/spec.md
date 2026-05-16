## ADDED Requirements
### Requirement: Session list CLI output selection stays explicit
The session invalidation CLI documentation SHALL describe the session listing command with `--output table|json` instead of a boolean `--json` toggle.

#### Scenario: Session list docs use explicit output format
- **WHEN** a user reads the session invalidation CLI summary
- **THEN** they see `lango p2p session list [--output table|json]`
