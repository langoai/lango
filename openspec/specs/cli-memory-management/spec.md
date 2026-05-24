## Purpose

Capability spec for cli-memory-management. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Memory command output routing
The system SHALL route `lango memory list` and `lango memory status` output through the Cobra command writer instead of writing directly to process stdout.

#### Scenario: Memory output uses the command writer
- **WHEN** `lango memory list` or `lango memory status` renders human-readable or JSON output
- **THEN** the command SHALL write the full output through the Cobra command output writer
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` SHALL capture the command output

### Requirement: Memory list command
The system SHALL provide a `lango memory list --session <key> [--output table|json]` command that lists observations and reflections for a given session. The command SHALL support `--type observations|reflections` to filter by entry type. The `--session` flag SHALL be required. Table output SHALL display ID (truncated to 8 chars), TYPE, TOKENS, CREATED timestamp, and CONTENT (truncated to 60 characters).

#### Scenario: List all entries for a session
- **WHEN** user runs `lango memory list --session my-session`
- **THEN** the command displays a table of all observations and reflections for that session

#### Scenario: Filter by type
- **WHEN** user runs `lango memory list --session my-session --type observations`
- **THEN** the command displays only observations, excluding reflections

#### Scenario: JSON output
- **WHEN** user runs `lango memory list --session my-session --output json`
- **THEN** the command outputs a JSON array with id, type, tokens, created_at, and content fields

#### Scenario: Empty session JSON output
- **WHEN** user runs `lango memory list --session nonexistent --output json`
- **THEN** the command outputs `[]`

#### Scenario: Memory list rejects unknown output before config load
- **WHEN** user runs `lango memory list --session my-session --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

#### Scenario: Empty session
- **WHEN** user runs `lango memory list --session nonexistent`
- **THEN** the command displays "No entries found." and exits with code 0

### Requirement: Memory status command
The system SHALL provide a `lango memory status --session <key> [--output table|json]` command that displays observation and reflection counts, token totals, and Observational Memory configuration values. The `--session` flag SHALL be required.

#### Scenario: Display status
- **WHEN** user runs `lango memory status --session my-session`
- **THEN** the command displays enabled state, provider, model, observation/reflection counts with token totals, and threshold configuration values

#### Scenario: JSON status output
- **WHEN** user runs `lango memory status --session my-session --output json`
- **THEN** the command outputs a JSON object with observations, reflections, token counts, and configuration fields

#### Scenario: Memory status rejects unknown output before config load
- **WHEN** user runs `lango memory status --session my-session --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Memory clear command
The system SHALL provide a `lango memory clear <session-key>` command that deletes all observations and reflections for the given session. The session key SHALL be a positional argument. The command SHALL prompt for confirmation before deletion. The `--force` flag SHALL skip the confirmation prompt.

#### Scenario: Memory clear prompt uses command streams
- **WHEN** `lango memory clear` prompts for confirmation
- **THEN** it SHALL write the prompt through the Cobra command output writer
- **AND** it SHALL read the operator response through the Cobra command input reader
- **AND** wrappers or tests that replace `cmd.OutOrStdout()` and `cmd.InOrStdin()` SHALL control the interaction

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

### Requirement: ListAgentNames method on Store interface
The agentmemory.Store interface SHALL include a `ListAgentNames() ([]string, error)` method that returns the names of all agents that have stored memories. This method is required to support the `memory agents` CLI command.

#### Scenario: ListAgentNames with entries
- **WHEN** ListAgentNames() is called on a store containing entries for agents "researcher" and "planner"
- **THEN** the method returns ["researcher", "planner"] (order not guaranteed) with no error

#### Scenario: ListAgentNames with no entries
- **WHEN** ListAgentNames() is called on a store with no agent memory entries
- **THEN** the method returns an empty slice with no error

### Requirement: ListAll method on Store interface
The agentmemory.Store interface SHALL include a `ListAll(agentName string) ([]*Entry, error)` method that returns all memory entries for the specified agent. This method is required to support the `memory agent <name>` CLI command.

#### Scenario: ListAll for existing agent
- **WHEN** ListAll("researcher") is called on a store containing 5 entries for "researcher"
- **THEN** the method returns all 5 Entry pointers with no error

#### Scenario: ListAll for nonexistent agent
- **WHEN** ListAll("unknown") is called on a store with no entries for "unknown"
- **THEN** the method returns an empty slice with no error

### Requirement: MemStore implementation
The in-memory agentmemory.MemStore implementation SHALL implement both ListAgentNames() and ListAll() by iterating the internal memory map.

#### Scenario: MemStore ListAgentNames
- **WHEN** ListAgentNames() is called on a MemStore with entries for 3 agents
- **THEN** the method returns a slice of 3 agent name strings

### Requirement: Backward compatibility
The addition of ListAgentNames() and ListAll() to the Store interface SHALL NOT change the behavior of existing Store methods. All existing tests SHALL continue to pass.

#### Scenario: Existing tests pass
- **WHEN** `go test ./internal/agentmemory/...` is run after the interface additions
- **THEN** all existing tests pass without modification

### Requirement: Agent memory agents command
The system SHALL provide a `lango memory agents [--output table|json]` command that lists agents with persistent memory entries from the configured session database.

#### Scenario: Agent memory agents table output
- **WHEN** user runs `lango memory agents`
- **THEN** the command outputs a table with AGENT, ENTRIES, and LAST UPDATED columns

#### Scenario: Agent memory agents JSON output
- **WHEN** user runs `lango memory agents --output json`
- **THEN** the command outputs a JSON array of agent summary objects

#### Scenario: Agent memory agents reject unknown output before config load
- **WHEN** user runs `lango memory agents --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader

### Requirement: Agent memory single-agent command
The system SHALL provide a `lango memory agent <name> [--limit N] [--output table|json]` command that lists stored memory entries for the specified agent from the configured session database.

#### Scenario: Agent memory entry table output
- **WHEN** user runs `lango memory agent <name>`
- **THEN** the command outputs a table with ID, CONTENT, USES, and CREATED columns

#### Scenario: Agent memory command honors limit
- **WHEN** user runs `lango memory agent <name> --limit 1`
- **THEN** the command outputs at most one stored entry

#### Scenario: Agent memory command JSON output
- **WHEN** user runs `lango memory agent <name> --output json`
- **THEN** the command outputs a JSON array of stored memory entries

#### Scenario: Agent memory command rejects unknown output before config load
- **WHEN** user runs `lango memory agent <name> --output yaml`
- **THEN** the command SHALL return an actionable unknown-output-format error
- **AND** it SHALL NOT invoke the config loader
