## MODIFIED Requirements

### Requirement: CLI system metrics summary
The CLI SHALL provide `lango metrics` that fetches the `/metrics` JSON endpoint and renders a system metrics snapshot summary as table (default) or JSON. When `--addr` is omitted, the command SHALL resolve the gateway address from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: Metrics summary table output
- **WHEN** `lango metrics` is run without --output flag
- **THEN** it SHALL fetch `/metrics` and render uptime, token usage, and tool execution counts in table format

#### Scenario: Metrics summary JSON output
- **WHEN** `lango metrics --output json` is run
- **THEN** it SHALL output raw JSON from the `/metrics` endpoint

#### Scenario: Metrics summary rejects an unknown output format before fetch
- **WHEN** `lango metrics --output yaml` is run
- **THEN** the command SHALL return an error before contacting the gateway

#### Scenario: Metrics summary uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs `lango metrics` without `--addr`
- **THEN** the command SHALL fetch `/metrics` from that configured gateway address

#### Scenario: Metrics summary explicit address override
- **WHEN** the user runs `lango metrics --addr <url>`
- **THEN** the command SHALL fetch `/metrics` from `<url>` instead of the configured gateway address

### Requirement: CLI metrics breakdown subcommands
The CLI SHALL provide `lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history` that fetch their respective `/metrics/*` endpoints and render table (default) or JSON output. When `--addr` is omitted, each subcommand SHALL resolve the gateway address from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: Sessions metrics
- **WHEN** `lango metrics sessions` is run
- **THEN** it SHALL fetch `/metrics/sessions` and display per-session token usage

#### Scenario: Tool metrics
- **WHEN** `lango metrics tools` is run
- **THEN** it SHALL fetch `/metrics/tools` and display tool execution statistics

#### Scenario: Agent metrics
- **WHEN** `lango metrics agents` is run
- **THEN** it SHALL fetch `/metrics/agents` and display per-agent token usage

#### Scenario: Historical metrics
- **WHEN** `lango metrics history --days 7` is run
- **THEN** it SHALL fetch `/metrics/history?days=7` and display historical token usage

#### Scenario: Metrics breakdown uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs a metrics breakdown subcommand without `--addr`
- **THEN** the command SHALL fetch its gateway endpoint from that configured gateway address
