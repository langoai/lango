## MODIFIED Requirements

### Requirement: Security status command
Security status reads SHALL use broker-backed storage diagnostics rather than opening the SQLite database directly from the CLI process. The status surface SHALL also report whether the first-slice exportability policy is enabled.

#### Scenario: Status command reads through broker
- **WHEN** the security status command needs database-backed counts or metadata
- **THEN** it SHALL query the broker-backed storage layer instead of opening the database directly in the CLI process

#### Scenario: Exportability status reported
- **WHEN** the user runs the security status command
- **THEN** the output SHALL include whether exportability evaluation is enabled in the active config
