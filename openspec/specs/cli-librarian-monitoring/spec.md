# CLI Librarian Monitoring

## Purpose
Provides CLI commands for monitoring the librarian system, including viewing configuration status and browsing inquiry history.
## Requirements
### Requirement: Librarian status command
The system SHALL provide a `lango librarian status [--json]` command that displays the current librarian configuration including enabled state, knowledge sources, and indexing settings. The command SHALL use cfgLoader (config only).

#### Scenario: Librarian enabled
- **WHEN** user runs `lango librarian status` with librarian enabled
- **THEN** system displays enabled state, configured knowledge sources, and inquiry handling mode

#### Scenario: Librarian disabled
- **WHEN** user runs `lango librarian status` with librarian disabled
- **THEN** system displays "Librarian is disabled"

#### Scenario: Librarian status in JSON format
- **WHEN** user runs `lango librarian status --json`
- **THEN** system outputs a JSON object with fields: enabled, knowledgeSources, inquiryMode

### Requirement: Librarian inquiries command
The system SHALL provide a `lango librarian inquiries [--limit N] [--json]` command that displays recent librarian inquiry records from the database. The command SHALL use bootLoader because it requires database access. The default limit SHALL be 20 entries.

#### Scenario: Inquiries with default limit
- **WHEN** user runs `lango librarian inquiries`
- **THEN** system displays up to 20 most recent inquiries in a table with TIMESTAMP, QUERY, and STATUS columns

#### Scenario: Inquiries with custom limit
- **WHEN** user runs `lango librarian inquiries --limit 10`
- **THEN** system displays up to 10 most recent inquiries

#### Scenario: No inquiries recorded
- **WHEN** user runs `lango librarian inquiries` with no inquiry history
- **THEN** system displays "No librarian inquiries found"

#### Scenario: Inquiries in JSON format
- **WHEN** user runs `lango librarian inquiries --json`
- **THEN** system outputs a JSON array of inquiry objects

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

