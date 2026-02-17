## ADDED Requirements

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with specialized sub-agents: Executor, Researcher, Planner, and MemoryManager. The orchestrator SHALL have NO direct tools (`Tools: nil`) and MUST delegate all tool-requiring tasks to sub-agents.

#### Scenario: Multi-agent mode enabled
- **WHEN** `agent.multiAgent` is true
- **THEN** BuildAgentTree SHALL create an orchestrator that has NO direct tools AND has sub-agents (Executor, Researcher, Planner, and MemoryManager)

#### Scenario: Orchestrator has no direct tools
- **WHEN** the orchestrator is created
- **THEN** the orchestrator's `Tools` field SHALL be `nil`
- **AND** tools SHALL only be adapted for their respective sub-agents (each tool adapted exactly once)

#### Scenario: Single-agent fallback
- **WHEN** `agent.multiAgent` is false
- **THEN** the system SHALL create a single flat agent with all tools

### Requirement: Tool partitioning by prefix
Tools SHALL be partitioned to sub-agents based on name prefixes: `exec/fs_/browser_/crypto_/skill_/payment_` → Executor, `search_/rag_/graph_/save_knowledge/save_learning` → Researcher, `memory_/observe_/reflect_` → MemoryManager, unmatched → Executor.

#### Scenario: Graph tools routed to Researcher
- **WHEN** tools named `graph_traverse` and `graph_query` are registered
- **THEN** they SHALL be assigned to the Researcher sub-agent

#### Scenario: Unmatched tools default to Executor
- **WHEN** a tool with an unrecognized prefix is present
- **THEN** it SHALL be assigned to the Executor sub-agent

#### Scenario: Payment tools routed to Executor
- **WHEN** tools with prefix `payment_` are registered
- **THEN** they SHALL be assigned to the Executor sub-agent
- **AND** the capabilityMap SHALL describe them as "blockchain payments (USDC on Base)"

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

### Requirement: Capability-based sub-agent descriptions
Sub-agent descriptions in the orchestrator prompt SHALL use human-readable capability summaries instead of raw tool names. A `capabilityMap` SHALL map tool name prefixes to natural-language descriptions (e.g., `browser_` → "web browsing", `exec` → "command execution"). The `capabilityDescription()` function SHALL deduplicate capabilities across a tool set.

#### Scenario: Executor description uses capabilities
- **WHEN** the executor sub-agent has tools `exec_shell`, `fs_read`, `browser_navigate`
- **THEN** its description SHALL contain "command execution, file operations, web browsing"
- **AND** it SHALL NOT contain raw tool names like "exec_shell" or "browser_navigate"

#### Scenario: Duplicate capabilities are deduplicated
- **WHEN** two tools share the same prefix (e.g., `exec_shell` and `exec_run`)
- **THEN** the capability "command execution" SHALL appear only once in the description

#### Scenario: Unknown tool prefix falls back to general actions
- **WHEN** a tool has no matching prefix in `capabilityMap`
- **THEN** its capability SHALL be "general actions"

### Requirement: Orchestrator instruction guides delegation-only execution
The orchestrator instruction SHALL enforce mandatory delegation for all tool-requiring tasks. It SHALL list all valid sub-agent names with exact spelling, explicitly prohibit inventing or abbreviating agent names, and instruct the LLM that it has no tools of its own. The instruction SHALL NOT contain words that could be confused with agent names (e.g., avoid listing action types like "browser" that might be mistaken for an agent name). Sub-agent entries in the instruction SHALL use capability descriptions, not raw tool name lists.

#### Scenario: Tool-requiring task
- **WHEN** a user requests any task requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent using its exact registered name

#### Scenario: Agent name exactness
- **WHEN** the orchestrator delegates to a sub-agent
- **THEN** it SHALL use the EXACT name (e.g. "executor", NOT "exec", "browser", or any abbreviation)

#### Scenario: Delegation rules by sub-agent role
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL specify: executor for actions, researcher for information lookup, planner for multi-step planning, memory-manager for memory operations

#### Scenario: Invalid agent name prevention
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL contain the text "NEVER invent or abbreviate agent names"
- **AND** it SHALL list only the exact names of registered sub-agents

#### Scenario: Sub-agent descriptions use capabilities not tool names
- **WHEN** the orchestrator instruction lists sub-agents
- **THEN** each sub-agent entry SHALL describe capabilities (e.g., "command execution, file operations")
- **AND** SHALL NOT contain raw tool names (e.g., "exec_shell", "browser_navigate")

### Requirement: Orchestrator system prompt isolation
The orchestrator system prompt SHALL NOT include tool-category descriptions (SectionIdentity from AGENTS.md) or tool-usage guidelines (SectionToolUsage from TOOL_USAGE.md). These sections reference tool names like "Exec", "Browser", "Crypto" that the LLM may misinterpret as agent names.

#### Scenario: Orchestrator prompt construction
- **WHEN** multi-agent mode is enabled
- **THEN** the orchestrator prompt SHALL replace SectionIdentity with a delegation-focused identity
- **AND** the orchestrator prompt SHALL remove SectionToolUsage entirely

#### Scenario: Single-agent prompt unaffected
- **WHEN** multi-agent mode is disabled
- **THEN** the single-agent prompt SHALL retain all sections including SectionIdentity and SectionToolUsage

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
The orchestrator instruction SHALL direct the LLM to respond directly to simple conversational queries (greetings, opinions, general knowledge) without delegating to sub-agents.

#### Scenario: Simple greeting
- **WHEN** user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Task requiring tools
- **WHEN** user requests an action requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

### Requirement: Max Delegation Rounds
The `Config` struct SHALL include a `MaxDelegationRounds` field. The orchestrator instruction SHALL mention this limit as a prompt-based guardrail.

#### Scenario: Default max rounds
- **WHEN** `MaxDelegationRounds` is zero or unset
- **THEN** the default limit of 5 rounds SHALL be used in the orchestrator prompt

### Requirement: Dynamic Orchestrator Instruction
The orchestrator instruction SHALL be dynamically generated to list only the sub-agents that were actually created, rather than hardcoding all four agent names.

#### Scenario: Only executor and planner created
- **WHEN** only executor and planner sub-agents are created
- **THEN** the orchestrator instruction SHALL only mention executor and planner

### Requirement: Sub-Agent Result Reporting
Each sub-agent instruction SHALL include guidance to report results clearly after completing their task.

#### Scenario: Executor result reporting
- **WHEN** the executor sub-agent completes an action
- **THEN** its instruction SHALL guide it to provide results clearly

#### Scenario: Researcher result reporting
- **WHEN** the researcher sub-agent completes research
- **THEN** its instruction SHALL guide it to summarize findings clearly

#### Scenario: Planner result reporting
- **WHEN** the planner sub-agent completes planning
- **THEN** its instruction SHALL guide it to present the plan for review

#### Scenario: Memory Manager result reporting
- **WHEN** the memory-manager sub-agent completes memory operations
- **THEN** its instruction SHALL guide it to report what was stored or retrieved
