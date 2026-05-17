## MODIFIED Requirements

### Requirement: CLI subcommand for policy metrics
The CLI SHALL provide a `lango metrics policy` subcommand that fetches from `/metrics/policy` and renders as table (default) or JSON. When `--addr` is omitted, the command SHALL resolve the gateway address from configured `server.host` and `server.port`, falling back to `http://localhost:18789` only when configuration is unavailable, blank, or zero-valued.

#### Scenario: Policy metrics table output
- **WHEN** `lango metrics policy` is run without --output flag
- **THEN** it SHALL fetch `/metrics/policy` and render block/observe counts and counts by reason in table format

#### Scenario: Policy metrics JSON output
- **WHEN** `lango metrics policy --output json` is run
- **THEN** it SHALL fetch `/metrics/policy` and output the raw policy metrics JSON

#### Scenario: Policy metrics uses configured gateway by default
- **WHEN** configuration sets `server.host` and `server.port`
- **AND** the user runs `lango metrics policy` without `--addr`
- **THEN** the command SHALL fetch `/metrics/policy` from that configured gateway address
