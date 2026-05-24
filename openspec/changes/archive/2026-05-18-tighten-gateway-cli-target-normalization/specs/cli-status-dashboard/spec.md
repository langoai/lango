## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view. When `--addr` is supplied, the live server probe SHALL use the normalized explicit address and the status output SHALL report that same normalized gateway target.

#### Scenario: Status explicit address display matches probe target
- **WHEN** the user runs `lango status --addr <url>`
- **THEN** the command SHALL probe `/health` on the normalized `<url>`
- **AND** the displayed gateway field SHALL match the same normalized `<url>`

#### Scenario: Status explicit address trims trailing slash
- **WHEN** the user runs `lango status --addr http://127.0.0.1:18789/`
- **THEN** the command SHALL probe `/health` on `http://127.0.0.1:18789`
- **AND** the displayed gateway field SHALL be `http://127.0.0.1:18789`
