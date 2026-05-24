# CLI Learning Inspection

## Purpose
Provides CLI commands for inspecting the learning system configuration and viewing learning history audit logs.
## Requirements
### Requirement: Learning command output routing
The system SHALL route `lango learning status` and `lango learning history` output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Learning output uses the command writer
- **WHEN** `lango learning status` or `lango learning history` renders human-readable or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Learning status command
The system SHALL provide a `lango learning status [--output table|json]` command that displays the current learning system configuration including knowledge settings, graph settings, and embedding/RAG settings. The command SHALL use cfgLoader (config only).

#### Scenario: Learning enabled
- **WHEN** user runs `lango learning status` with learning system enabled
- **THEN** system displays knowledge enabled state, error correction state, confidence threshold, graph state, and embedding/RAG settings

#### Scenario: Learning disabled
- **WHEN** user runs `lango learning status` with learning disabled
- **THEN** system still renders the status sections with disabled boolean values reflected in the output

#### Scenario: Learning status in JSON format
- **WHEN** user runs `lango learning status --output json`
- **THEN** system outputs a JSON object containing the current knowledge, graph, embedding, and RAG fields

#### Scenario: Learning status rejects unknown output before config load
- **WHEN** user runs `lango learning status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Learning history command
The system SHALL provide a `lango learning history [--limit N] [--output table|json]` command that displays recent learning audit log entries from the database. The command SHALL use bootLoader because it requires database access. The default limit SHALL be 20 entries.

#### Scenario: History with default limit
- **WHEN** user runs `lango learning history`
- **THEN** system displays up to 20 most recent learning events in a table with ID, CATEGORY, TRIGGER, CONFIDENCE, and CREATED columns

#### Scenario: History with custom limit
- **WHEN** user runs `lango learning history --limit 5`
- **THEN** system displays up to 5 most recent learning events

#### Scenario: Empty history
- **WHEN** user runs `lango learning history` with no learning events recorded
- **THEN** system displays "No learning entries found."

#### Scenario: History in JSON format
- **WHEN** user runs `lango learning history --output json`
- **THEN** system outputs a JSON array of learning event objects

#### Scenario: History rejects unknown output before bootstrap
- **WHEN** user runs `lango learning history --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the bootstrap loader

### Requirement: Learning command group entry
The system SHALL provide a `lango learning` command group that shows help text listing status and history subcommands.

#### Scenario: Help text
- **WHEN** user runs `lango learning`
- **THEN** system displays help listing status and history subcommands

### Requirement: Learning history uses storage reader
The `lango learning history` command MUST read recent learning rows through a storage facade reader instead of querying Ent directly from the CLI layer.

#### Scenario: Learning history command reads through facade
- **WHEN** the user runs `lango learning history`
- **THEN** the command loads recent learning records from the storage facade reader

### Requirement: Learning history supports broker-backed runtime reads
The `lango learning history` command MUST remain functional when bootstrap is broker-owned and runtime reads come from broker-backed storage.

#### Scenario: Learning history under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** `lango learning history` still returns recent learning records through the broker-backed reader path
