## Purpose

Capability spec for cli-memory-management. See requirements below for scope and behavior contracts.

## Requirements

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
