## MODIFIED Requirements

### Requirement: README reflects all implemented features
The README SHALL list all implemented features including Team Health Monitoring,
Incremental Git Bundles, Task Branch Management, Config Presets,
Event-Driven Bridges, EventMonitor Reorg Protection, and Escrow Hub V2.

#### Scenario: CLI commands in README
- **WHEN** a user reads the CLI commands section of `README.md`
- **THEN** `lango status`, `lango onboard --preset`, cron `--timeout`, cron
  `--deliver`, and cron management by `id-or-name` SHALL be documented

### Requirement: Public cron automation docs match CLI flags
Public cron automation documentation SHALL use command examples accepted by the
current CLI.

#### Scenario: Cron docs show accepted add flags
- **WHEN** a user reads `docs/automation/cron.md`
- **THEN** the add examples SHALL use accepted delivery flags
- **AND** per-job timeout examples SHALL match `lango cron add --timeout`

#### Scenario: Cron docs show accepted control selectors
- **WHEN** a user reads `docs/automation/cron.md`
- **THEN** pause, resume, delete, and job-specific history examples SHALL use
  accepted cron job selectors
