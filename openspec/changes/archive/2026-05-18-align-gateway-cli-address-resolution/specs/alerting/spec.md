## MODIFIED Requirements

### Requirement: CLI alerts list command
The system SHALL provide a `lango alerts list` CLI command that queries the gateway `/alerts` endpoint and displays recent alerts. The command SHALL support `--days` flag (default: 7) and `--output table|json` format. When `--addr` is omitted, the command SHALL resolve the gateway address from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: List alerts in table format
- **WHEN** user runs `lango alerts list`
- **THEN** the system displays alerts from the last 7 days in table format with columns: time, type, severity, message

#### Scenario: List alerts in JSON format
- **WHEN** user runs `lango alerts list --output json`
- **THEN** the system outputs alerts as a JSON array

#### Scenario: Alerts list rejects an unknown output format before fetch
- **WHEN** user runs `lango alerts list --output yaml`
- **THEN** the system returns an actionable unknown-output-format error
- **AND** it SHALL NOT contact the gateway

#### Scenario: Alerts list uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs `lango alerts list` without `--addr`
- **THEN** the command SHALL fetch `/alerts` from that configured gateway address

#### Scenario: Alerts command output uses the command writer
- **WHEN** `lango alerts list` or `lango alerts summary` renders table or JSON output
- **THEN** it SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: CLI alerts summary command
The system SHALL provide a `lango alerts summary` CLI command that queries the gateway `/alerts` endpoint and displays aggregated alert counts by type. When `--addr` is omitted, the command SHALL resolve the gateway address from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: Summary with alerts
- **WHEN** user runs `lango alerts summary` and alerts exist
- **THEN** the system displays a count of alerts grouped by type

#### Scenario: Alerts summary uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs `lango alerts summary` without `--addr`
- **THEN** the command SHALL fetch `/alerts` from that configured gateway address

#### Scenario: Alerts explicit address override
- **WHEN** the user runs `lango alerts list --addr <url>` or `lango alerts summary --addr <url>`
- **THEN** the command SHALL fetch from `<url>` instead of the configured gateway address
