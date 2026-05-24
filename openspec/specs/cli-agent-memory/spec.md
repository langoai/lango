# CLI Agent Memory

## Purpose
Provides CLI commands for inspecting agent memory entries, including listing agents with stored memories and viewing detailed memory entries for a specific agent.
## Requirements
### Requirement: Memory agents command
The system SHALL provide a `lango memory agents [--output table|json]` command that lists all agent names that have stored memories by calling ListAgentNames() on the agentmemory.Store interface. The command SHALL use bootLoader because it requires database access.

#### Scenario: Agents with memories
- **WHEN** user runs `lango memory agents`
- **THEN** system displays a list of agent names that have stored memory entries

#### Scenario: No agents with memories
- **WHEN** user runs `lango memory agents` with no agent memory entries
- **THEN** system displays "No agent memories found"

#### Scenario: Agents list in JSON format
- **WHEN** user runs `lango memory agents --output json`
- **THEN** system outputs a JSON array of agent name strings

### Requirement: Memory agent detail command
The system SHALL provide a `lango memory agent <name> [--output table|json]` command that lists all memory entries for a specific agent by calling ListAll(agentName) on the agentmemory.Store interface. Each entry SHALL display key, scope, kind, confidence, use count, and content preview.

#### Scenario: Agent has memories
- **WHEN** user runs `lango memory agent researcher`
- **THEN** system displays a table with KEY, SCOPE, KIND, CONFIDENCE, USE COUNT, and CONTENT columns for all entries belonging to "researcher"

#### Scenario: Agent has no memories
- **WHEN** user runs `lango memory agent unknown-agent`
- **THEN** system displays "No memories found for agent 'unknown-agent'"

#### Scenario: Agent detail in JSON format
- **WHEN** user runs `lango memory agent researcher --output json`
- **THEN** system outputs a JSON array of Entry objects with id, agent_name, scope, kind, key, content, confidence, use_count, tags, created_at, and updated_at fields

### Requirement: Memory agent commands registration
The `agents` and `agent` subcommands SHALL be registered under the existing `lango memory` command group.

#### Scenario: Memory help lists new subcommands
- **WHEN** user runs `lango memory --help`
- **THEN** the help output includes agents and agent alongside existing subcommands

### Requirement: Memory clear confirmation uses shared command streams
`lango memory clear <session-key>` SHALL drive its confirmation prompt through the shared confirmation helper using Cobra command input/output streams.

#### Scenario: Memory clear aborts on denial
- **WHEN** `lango memory clear user-123` prompts for confirmation and the user answers `n`
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave memory entries untouched

#### Scenario: Memory clear prompt uses command streams
- **WHEN** `lango memory clear user-123` prompts for confirmation
- **THEN** the warning line and `Continue? [y/N]: ` prompt SHALL be written through the Cobra command output stream
- **AND** the operator response SHALL be read through the Cobra command input stream

### Requirement: Memory clear treats EOF as denial
`lango memory clear <session-key>` SHALL treat EOF on its confirmation input as a clean denial.

#### Scenario: Memory clear EOF aborts cleanly
- **WHEN** `lango memory clear user-123` prompts for confirmation and stdin reaches EOF before approval
- **THEN** the command SHALL print `Aborted.`
- **AND** it SHALL leave memory entries untouched
