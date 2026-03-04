## ADDED Requirements

### Requirement: Learning status command
The system SHALL provide a `lango learning status [--json]` command that displays the current learning system configuration including enabled state, graph engine settings, and confidence propagation rate. The command SHALL use cfgLoader (config only).

#### Scenario: Learning enabled
- **WHEN** user runs `lango learning status` with learning system enabled
- **THEN** system displays enabled state, graph engine status, confidence propagation rate, and auto-learn setting

#### Scenario: Learning disabled
- **WHEN** user runs `lango learning status` with learning disabled
- **THEN** system displays "Learning system is disabled"

#### Scenario: Learning status in JSON format
- **WHEN** user runs `lango learning status --json`
- **THEN** system outputs a JSON object with fields: enabled, graphEngine, confidencePropagationRate, autoLearn

### Requirement: Learning history command
The system SHALL provide a `lango learning history [--limit N] [--json]` command that displays recent learning audit log entries from the database. The command SHALL use bootLoader because it requires database access. The default limit SHALL be 20 entries.

#### Scenario: History with default limit
- **WHEN** user runs `lango learning history`
- **THEN** system displays up to 20 most recent learning events in a table with TIMESTAMP, TYPE, and SUMMARY columns

#### Scenario: History with custom limit
- **WHEN** user runs `lango learning history --limit 5`
- **THEN** system displays up to 5 most recent learning events

#### Scenario: Empty history
- **WHEN** user runs `lango learning history` with no learning events recorded
- **THEN** system displays "No learning history found"

#### Scenario: History in JSON format
- **WHEN** user runs `lango learning history --json`
- **THEN** system outputs a JSON array of learning event objects

### Requirement: Learning command group entry
The system SHALL provide a `lango learning` command group that shows help text listing status and history subcommands.

#### Scenario: Help text
- **WHEN** user runs `lango learning`
- **THEN** system displays help listing status and history subcommands
