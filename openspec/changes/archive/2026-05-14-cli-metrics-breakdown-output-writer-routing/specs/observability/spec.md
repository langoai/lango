## ADDED Requirements

### Requirement: CLI metrics breakdown subcommands

The CLI SHALL provide `lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history` that fetch their respective `/metrics/*` endpoints and render table (default) or JSON output.

#### Scenario: Sessions output
- **WHEN** `lango metrics sessions` is run
- **THEN** it SHALL display per-session token usage or an empty-state message when no session data exists

#### Scenario: Tools output
- **WHEN** `lango metrics tools` is run
- **THEN** it SHALL display per-tool execution statistics or an empty-state message when no tool data exists

#### Scenario: Agents output
- **WHEN** `lango metrics agents` is run
- **THEN** it SHALL display per-agent token usage or an empty-state message when no agent data exists

#### Scenario: History output
- **WHEN** `lango metrics history --days 7` is run
- **THEN** it SHALL display historical token usage records and aggregate totals or an empty-state message when no history exists

### Requirement: CLI metrics breakdown output routing

`lango metrics sessions`, `lango metrics tools`, `lango metrics agents`, and `lango metrics history` SHALL write all non-error output through the Cobra command output stream so wrappers and test harnesses can capture table and JSON output without intercepting process-global stdout.

#### Scenario: Sessions JSON writes to command output
- **WHEN** `lango metrics sessions --output json` is run
- **THEN** the command writes the JSON payload to the Cobra command output stream

#### Scenario: Tools empty-state writes to command output
- **WHEN** `lango metrics tools` returns no data
- **THEN** the command writes the empty-state message to the Cobra command output stream

#### Scenario: Agents empty-state writes to command output
- **WHEN** `lango metrics agents` returns no data
- **THEN** the command writes the empty-state message to the Cobra command output stream

#### Scenario: History table writes to command output
- **WHEN** `lango metrics history --days 3` is run
- **THEN** the command writes the history summary and table to the Cobra command output stream
