## MODIFIED Requirements

### Requirement: Unified status dashboard command
The system SHALL provide a `lango status` command that displays system health, configuration state, active channels, and feature status in a single dashboard view. When `--addr` is omitted, the live server probe SHALL use the gateway address resolved from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: Server not running
- **WHEN** user runs `lango status` and the server is not running
- **THEN** system displays config-based status (profile, gateway address, provider, features) with server marked as "not running"

#### Scenario: Server running
- **WHEN** user runs `lango status` and the server is running
- **THEN** system displays live health data alongside config-based status with server marked as "running"

#### Scenario: JSON output
- **WHEN** user runs `lango status --output json`
- **THEN** system outputs all status data as a JSON object with version, profile, serverUp, gateway, provider, model, features, channels, and serverInfo fields

#### Scenario: Status probe uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs `lango status` without `--addr`
- **THEN** the command SHALL probe `/health` on that configured gateway address
- **AND** the displayed gateway field SHALL match the same configured gateway address

#### Scenario: Status explicit address override
- **WHEN** the user runs `lango status --addr <url>`
- **THEN** the command SHALL probe `/health` on `<url>` instead of the configured gateway address
