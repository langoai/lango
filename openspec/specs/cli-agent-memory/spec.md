# CLI Agent Memory

## Purpose
Provides CLI commands for inspecting agent memory entries, including listing agents with stored memories and viewing detailed memory entries for a specific agent.

## Requirements

### Requirement: Memory agents command
The system SHALL provide a `lango memory agents [--json]` command that lists all agent names that have stored memories by calling ListAgentNames() on the agentmemory.Store interface. The command SHALL use bootLoader because it requires database access.

#### Scenario: Agents with memories
- **WHEN** user runs `lango memory agents`
- **THEN** system displays a list of agent names that have stored memory entries

#### Scenario: No agents with memories
- **WHEN** user runs `lango memory agents` with no agent memory entries
- **THEN** system displays "No agent memories found"

#### Scenario: Agents list in JSON format
- **WHEN** user runs `lango memory agents --json`
- **THEN** system outputs a JSON array of agent name strings

### Requirement: Memory agent detail command
The system SHALL provide a `lango memory agent <name> [--json]` command that lists all memory entries for a specific agent by calling ListAll(agentName) on the agentmemory.Store interface. Each entry SHALL display key, scope, kind, confidence, use count, and content preview.

#### Scenario: Agent has memories
- **WHEN** user runs `lango memory agent researcher`
- **THEN** system displays a table with KEY, SCOPE, KIND, CONFIDENCE, USE COUNT, and CONTENT columns for all entries belonging to "researcher"

#### Scenario: Agent has no memories
- **WHEN** user runs `lango memory agent unknown-agent`
- **THEN** system displays "No memories found for agent 'unknown-agent'"

#### Scenario: Agent detail in JSON format
- **WHEN** user runs `lango memory agent researcher --json`
- **THEN** system outputs a JSON array of Entry objects with id, agent_name, scope, kind, key, content, confidence, use_count, tags, created_at, and updated_at fields

### Requirement: Memory agent commands registration
The `agents` and `agent` subcommands SHALL be registered under the existing `lango memory` command group.

#### Scenario: Memory help lists new subcommands
- **WHEN** user runs `lango memory --help`
- **THEN** the help output includes agents and agent alongside existing subcommands
