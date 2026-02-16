## MODIFIED Requirements

### Requirement: Event Author Identity
The EventsAdapter SHALL use the stored `msg.Author` when available, falling back to the `rootAgentName` for assistant messages when no stored author exists. The author SHALL NOT be hardcoded to a fixed agent name.

#### Scenario: Multi-agent mode with stored author
- **WHEN** a message has `Author: "lango-orchestrator"` stored in history
- **THEN** the EventsAdapter SHALL use `"lango-orchestrator"` as the event author

#### Scenario: Multi-agent mode without stored author (legacy messages)
- **WHEN** a message has no stored Author and role is "assistant"
- **THEN** the EventsAdapter SHALL use the configured `rootAgentName` as the event author

#### Scenario: Single-agent mode
- **WHEN** the agent is created via `NewAgent()` (single-agent mode)
- **THEN** the rootAgentName SHALL be `"lango-agent"` and used for assistant events

### Requirement: Conditional Sub-Agent Creation
The `BuildAgentTree` function SHALL only create sub-agents that have tools assigned by `PartitionTools`. The Planner sub-agent SHALL always be created as it is LLM-only.

#### Scenario: All tool categories have tools
- **WHEN** tools exist for executor, researcher, and memory-manager roles
- **THEN** all four sub-agents (executor, researcher, planner, memory-manager) SHALL be created

#### Scenario: No memory tools assigned
- **WHEN** no tools match memory prefixes
- **THEN** the memory-manager sub-agent SHALL NOT be created

#### Scenario: No tools at all
- **WHEN** the tool list is empty
- **THEN** only the planner sub-agent SHALL be created

### Requirement: Orchestrator Short-Circuit for Simple Queries
The orchestrator instruction SHALL direct the LLM to respond directly to simple conversational queries (greetings, small talk, clarifying questions) without delegating to sub-agents.

#### Scenario: Simple greeting
- **WHEN** user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Complex task requiring tools
- **WHEN** user requests an action requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

### Requirement: Max Delegation Rounds
The `Config` struct SHALL include a `MaxDelegationRounds` field. The orchestrator instruction SHALL mention this limit as a prompt-based guardrail.

#### Scenario: Default max rounds
- **WHEN** `MaxDelegationRounds` is zero or unset
- **THEN** the default limit of 3 rounds SHALL be used in the orchestrator prompt

### Requirement: Dynamic Orchestrator Instruction
The orchestrator instruction SHALL be dynamically generated to list only the sub-agents that were actually created, rather than hardcoding all four agent names.

#### Scenario: Only executor and planner created
- **WHEN** only executor and planner sub-agents are created
- **THEN** the orchestrator instruction SHALL only mention executor and planner
