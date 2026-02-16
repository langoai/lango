## ADDED Requirements

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with specialized sub-agents: Executor, Researcher, Planner, and MemoryManager. The orchestrator SHALL also receive ALL tools directly via `llmagent.Config.Tools` so it can handle simple tasks without delegation.

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

### Requirement: Tool partitioning by prefix
Tools SHALL be partitioned to sub-agents based on name prefixes: `exec/fs_/browser_/crypto_/skill_` → Executor, `search_/rag_/graph_/save_knowledge/save_learning` → Researcher, `memory_/observe_/reflect_` → MemoryManager, unmatched → Executor.

#### Scenario: Graph tools routed to Researcher
- **WHEN** tools named `graph_traverse` and `graph_query` are registered
- **THEN** they SHALL be assigned to the Researcher sub-agent

#### Scenario: Unmatched tools default to Executor
- **WHEN** a tool with an unrecognized prefix is present
- **THEN** it SHALL be assigned to the Executor sub-agent

### Requirement: Graph, RAG, and Memory agent tools
The system SHALL provide dedicated tools for sub-agents: `graph_traverse`, `graph_query` (graph store), `rag_retrieve` (RAG service), `memory_list_observations`, `memory_list_reflections` (memory store).

#### Scenario: Graph tools available when graph enabled
- **WHEN** `graph.enabled: true`
- **THEN** `graph_traverse` and `graph_query` tools SHALL be added to the tool set

#### Scenario: RAG tool available when embedding configured
- **WHEN** embedding provider is configured and RAG service is initialized
- **THEN** `rag_retrieve` tool SHALL be added to the tool set

#### Scenario: Memory tools available when observational memory enabled
- **WHEN** `observationalMemory.enabled: true`
- **THEN** `memory_list_observations` and `memory_list_reflections` tools SHALL be added

### Requirement: Remote agents as sub-agents
The orchestrator SHALL accept remote A2A agents and append them to its sub-agent list. Remote agent names and descriptions SHALL be included in the orchestrator instruction.

#### Scenario: Remote agents loaded and wired
- **WHEN** `a2a.enabled: true` and `a2a.remoteAgents` contains entries
- **THEN** LoadRemoteAgents SHALL create ADK agents and they SHALL appear as sub-agents in the orchestrator

#### Scenario: Remote agent load failure
- **WHEN** a remote agent card URL is unreachable
- **THEN** the agent SHALL be skipped with a warning log, and the orchestrator SHALL continue with local sub-agents

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
