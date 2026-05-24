# CLI Librarian Monitoring

## Purpose
Provides CLI commands for monitoring the librarian system, including viewing configuration status and browsing inquiry history.
## Requirements
### Requirement: Librarian command output routing
The system SHALL route `lango librarian status` and `lango librarian inquiries` output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Librarian output uses the command writer
- **WHEN** `lango librarian status` or `lango librarian inquiries` renders human-readable or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Librarian status command
The system SHALL provide a `lango librarian status [--output table|json]` command that displays the current librarian configuration including enabled state, observation thresholds, inquiry cooldown, and configured model settings. The command SHALL use cfgLoader (config only).

#### Scenario: Librarian enabled
- **WHEN** user runs `lango librarian status` with librarian enabled
- **THEN** system displays enabled state, observation threshold, inquiry cooldown, pending inquiry limit, and auto-save confidence
- **AND** configured provider/model fields appear when set

#### Scenario: Librarian disabled
- **WHEN** user runs `lango librarian status` with librarian disabled
- **THEN** system still renders the status output with disabled boolean values reflected in the fields

#### Scenario: Librarian status in JSON format
- **WHEN** user runs `lango librarian status --output json`
- **THEN** system outputs a JSON object containing enabled state, thresholds, pending inquiry limit, auto-save confidence, and optional provider/model fields

#### Scenario: Librarian status rejects unknown output before config load
- **WHEN** user runs `lango librarian status --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Librarian inquiries command
The system SHALL provide a `lango librarian inquiries [--limit N] [--output table|json]` command that displays pending librarian inquiry records from the database. The command SHALL use bootLoader because it requires database access. The default limit SHALL be 20 entries.

#### Scenario: Inquiries with default limit
- **WHEN** user runs `lango librarian inquiries`
- **THEN** system displays up to 20 pending inquiries in a table with ID, PRIORITY, TOPIC, QUESTION, and CREATED columns

#### Scenario: Inquiries with custom limit
- **WHEN** user runs `lango librarian inquiries --limit 10`
- **THEN** system displays up to 10 most recent inquiries

#### Scenario: No inquiries recorded
- **WHEN** user runs `lango librarian inquiries` with no inquiry history
- **THEN** system displays "No pending inquiries."

#### Scenario: Inquiries in JSON format
- **WHEN** user runs `lango librarian inquiries --output json`
- **THEN** system outputs a JSON array of inquiry objects

#### Scenario: Inquiries rejects unknown output before bootstrap
- **WHEN** user runs `lango librarian inquiries --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the bootstrap loader

### Requirement: Librarian command group entry
The system SHALL provide a `lango librarian` command group that shows help text listing status and inquiries subcommands.

#### Scenario: Help text
- **WHEN** user runs `lango librarian`
- **THEN** system displays help listing status and inquiries subcommands

### Requirement: Librarian inquiries command uses storage reader
The `lango librarian inquiries` command MUST read pending inquiries through a storage facade reader instead of querying Ent directly from the CLI layer.

#### Scenario: Inquiries command reads through facade
- **WHEN** the user runs `lango librarian inquiries`
- **THEN** the command loads pending inquiry records from the storage facade reader

### Requirement: Librarian inquiries support broker-backed runtime reads
The `lango librarian inquiries` command MUST remain functional when bootstrap is broker-owned and runtime reads come from broker-backed storage.

#### Scenario: Librarian inquiries under broker-owned runtime
- **WHEN** broker-backed runtime storage is active
- **THEN** `lango librarian inquiries` still returns pending inquiry records through the broker-backed reader path
