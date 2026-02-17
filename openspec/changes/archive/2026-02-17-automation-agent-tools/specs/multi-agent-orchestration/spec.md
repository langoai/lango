## ADDED Requirements

### Requirement: Automator agent spec
The system SHALL include an "automator" `AgentSpec` in the `agentSpecs` registry for routing automation-related requests to a dedicated sub-agent.

#### Scenario: Automator routing
- **WHEN** tools with `cron_`, `bg_`, or `workflow_` prefixes are present
- **THEN** they SHALL be partitioned to the Automator role in `PartitionTools`

#### Scenario: Automator keywords
- **WHEN** a user request contains keywords like "schedule", "cron", "background", "workflow", "automate"
- **THEN** the orchestrator SHALL route to the automator sub-agent

### Requirement: Automator in RoleToolSet
The `RoleToolSet` SHALL include an `Automator []*agent.Tool` field, and `toolsForSpec` SHALL return it for the "automator" spec name.

#### Scenario: Tool partitioning order
- **WHEN** `PartitionTools` processes tools
- **THEN** automator matching SHALL occur before operator matching to prevent `cron_`/`bg_`/`workflow_` tools from being assigned to operator

### Requirement: Automation capability descriptions
The `capabilityMap` SHALL include entries for `cron_`, `bg_`, and `workflow_` prefixes.

#### Scenario: Capability description
- **WHEN** `toolCapability` is called for a `cron_` prefixed tool
- **THEN** it SHALL return "cron job scheduling"
