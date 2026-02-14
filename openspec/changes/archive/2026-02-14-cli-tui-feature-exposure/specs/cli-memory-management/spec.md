## ADDED Requirements

### Requirement: Memory list command
The system SHALL provide a `lango memory list --session <key>` command that lists observations and reflections for a given session. The command SHALL support `--type observations|reflections` to filter by entry type. The command SHALL support `--json` for JSON output. The `--session` flag SHALL be required. Table output SHALL display ID (truncated to 8 chars), TYPE, TOKENS, CREATED timestamp, and CONTENT (truncated to 60 characters).

#### Scenario: List all entries for a session
- **WHEN** user runs `lango memory list --session my-session`
- **THEN** the command displays a table of all observations and reflections for that session

#### Scenario: Filter by type
- **WHEN** user runs `lango memory list --session my-session --type observations`
- **THEN** the command displays only observations, excluding reflections

#### Scenario: JSON output
- **WHEN** user runs `lango memory list --session my-session --json`
- **THEN** the command outputs a JSON array with id, type, tokens, created_at, and content fields

#### Scenario: Empty session
- **WHEN** user runs `lango memory list --session nonexistent`
- **THEN** the command displays "No entries found." and exits with code 0

### Requirement: Memory status command
The system SHALL provide a `lango memory status --session <key>` command that displays observation and reflection counts, token totals, and Observational Memory configuration values. The `--session` flag SHALL be required. The command SHALL support `--json` for JSON output.

#### Scenario: Display status
- **WHEN** user runs `lango memory status --session my-session`
- **THEN** the command displays enabled state, provider, model, observation/reflection counts with token totals, and threshold configuration values

#### Scenario: JSON status output
- **WHEN** user runs `lango memory status --session my-session --json`
- **THEN** the command outputs a JSON object with observations, reflections, token counts, and configuration fields

### Requirement: Memory clear command
The system SHALL provide a `lango memory clear <session-key>` command that deletes all observations and reflections for the given session. The session key SHALL be a positional argument. The command SHALL prompt for confirmation before deletion. The `--force` flag SHALL skip the confirmation prompt.

#### Scenario: Clear with confirmation
- **WHEN** user runs `lango memory clear my-session` and confirms with "y"
- **THEN** the command deletes all observations and reflections for that session and displays a success message

#### Scenario: Clear aborted
- **WHEN** user runs `lango memory clear my-session` and answers "n"
- **THEN** the command displays "Aborted." and exits without deleting anything

#### Scenario: Force clear
- **WHEN** user runs `lango memory clear my-session --force`
- **THEN** the command deletes all entries without prompting for confirmation

### Requirement: Memory parent command
The system SHALL register `lango memory` as a top-level command with `list`, `status`, and `clear` subcommands. Running `lango memory` without a subcommand SHALL display help text listing available subcommands.

#### Scenario: Help output
- **WHEN** user runs `lango memory --help`
- **THEN** the command displays descriptions for list, status, and clear subcommands
