## MODIFIED Requirements

### Requirement: Hierarchical agent tree with 4 sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with 4 specialized sub-agents: Executor, Researcher, Planner, and MemoryManager. The orchestrator SHALL also receive ALL tools directly via `llmagent.Config.Tools` so it can handle simple tasks without delegation.

#### Scenario: Multi-agent mode enabled
- **WHEN** `agent.multiAgent` is true
- **THEN** BuildAgentTree SHALL create an orchestrator that has both direct tools AND sub-agents (Executor, Researcher, Planner, and MemoryManager)

#### Scenario: Orchestrator direct tool access
- **WHEN** the orchestrator is created with tools
- **THEN** all tools from `cfg.Tools` SHALL be adapted and assigned to the orchestrator's `Tools` field
- **AND** the same tools SHALL still be partitioned to their respective sub-agents

#### Scenario: Single-agent fallback
- **WHEN** `agent.multiAgent` is false
- **THEN** the system SHALL create a single flat agent with all tools

## ADDED Requirements

### Requirement: Orchestrator instruction guides direct vs delegated execution
The orchestrator instruction SHALL clearly distinguish between direct tool usage and sub-agent delegation. It SHALL list all valid sub-agent names and explicitly prohibit inventing agent names.

#### Scenario: Simple single-tool task
- **WHEN** a user requests a simple task requiring a single tool call
- **THEN** the orchestrator SHALL call the tool directly without delegating to a sub-agent

#### Scenario: Complex multi-step task
- **WHEN** a user requests a complex task requiring multiple steps or specialized reasoning
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent using its exact registered name

#### Scenario: Invalid agent name prevention
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL contain the text "NEVER invent agent names"
- **AND** it SHALL list only the exact names of registered sub-agents
