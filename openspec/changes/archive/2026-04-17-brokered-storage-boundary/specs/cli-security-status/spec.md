## MODIFIED Requirements

### Requirement: Security status command
Security status reads SHALL use broker-backed storage diagnostics rather than opening the SQLite database directly from the CLI process.

#### Scenario: Status command reads through broker
- **WHEN** the security status command needs database-backed counts or metadata
- **THEN** it SHALL query the broker-backed storage layer instead of opening the database directly in the CLI process
