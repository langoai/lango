## ADDED Requirements

### Requirement: Agent tools command
The system SHALL provide a `lango agent tools [--json]` command that lists all registered tools in the agent's tool catalog. The command SHALL use cfgLoader to load configuration and enumerate tools by name and description.

#### Scenario: List tools in text format
- **WHEN** user runs `lango agent tools`
- **THEN** system displays a table with NAME and DESCRIPTION columns for each registered tool

#### Scenario: List tools in JSON format
- **WHEN** user runs `lango agent tools --json`
- **THEN** system outputs a JSON array of tool objects with name and description fields

### Requirement: Agent hooks command
The system SHALL provide a `lango agent hooks [--json]` command that displays the current hook configuration including enabled hooks, blocked commands, and active hook types. The command SHALL use cfgLoader (config only).

#### Scenario: Hooks enabled
- **WHEN** user runs `lango agent hooks` with hooks.enabled set to true
- **THEN** system displays which hook types are active (securityFilter, accessControl, eventPublishing, knowledgeSave) and any blocked command patterns

#### Scenario: Hooks disabled
- **WHEN** user runs `lango agent hooks` with hooks.enabled set to false
- **THEN** system displays "Hooks are disabled"

#### Scenario: Hooks in JSON format
- **WHEN** user runs `lango agent hooks --json`
- **THEN** system outputs a JSON object with fields: enabled, securityFilter, accessControl, eventPublishing, knowledgeSave, blockedCommands

### Requirement: Agent command group registration
The `agent tools` and `agent hooks` subcommands SHALL be registered under the existing `lango agent` command group in `cmd/lango/main.go`.

#### Scenario: Agent help lists new subcommands
- **WHEN** user runs `lango agent --help`
- **THEN** the help output includes tools and hooks in the available subcommands list
