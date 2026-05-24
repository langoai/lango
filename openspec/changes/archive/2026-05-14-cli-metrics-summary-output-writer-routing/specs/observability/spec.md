## ADDED Requirements

### Requirement: CLI system metrics summary

The CLI SHALL provide `lango metrics` that fetches the `/metrics` JSON endpoint and renders a system metrics snapshot summary as table (default) or JSON.

#### Scenario: Metrics summary table output
- **WHEN** `lango metrics` is run without --output flag
- **THEN** it SHALL display uptime, total input/output token counts, and tool execution count

#### Scenario: Metrics summary JSON output
- **WHEN** `lango metrics --output json` is run
- **THEN** it SHALL output raw JSON from the `/metrics` endpoint

### Requirement: CLI system metrics output routing

`lango metrics` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Metrics summary table writes to command output
- **WHEN** `lango metrics` is run without --output flag
- **THEN** the command writes the table summary to the Cobra command output stream

#### Scenario: Metrics summary JSON writes to command output
- **WHEN** `lango metrics --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream
