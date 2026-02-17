## ADDED Requirements

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with specialized sub-agents: operator, navigator, vault, librarian, planner, and chronicler. The orchestrator SHALL have NO direct tools (`Tools: nil`) and MUST delegate all tool-requiring tasks to sub-agents.

#### Scenario: Multi-agent mode enabled
- **WHEN** `agent.multiAgent` is true
- **THEN** BuildAgentTree SHALL create an orchestrator that has NO direct tools AND has sub-agents (operator, navigator, vault, librarian, planner, chronicler)

#### Scenario: Orchestrator has no direct tools
- **WHEN** the orchestrator is created
- **THEN** the orchestrator's `Tools` field SHALL be `nil`
- **AND** tools SHALL only be adapted for their respective sub-agents (each tool adapted exactly once)

#### Scenario: Single-agent fallback
- **WHEN** `agent.multiAgent` is false
- **THEN** the system SHALL create a single flat agent with all tools

### Requirement: Tool partitioning by prefix
Tools SHALL be partitioned to sub-agents based on name prefixes with matching order Librarian → Chronicler → Navigator → Vault → Operator → Unmatched: `exec/fs_/skill_` → operator, `browser_` → navigator, `crypto_/secrets_/payment_` → vault, `search_/rag_/graph_/save_knowledge/save_learning/create_skill/list_skills` → librarian, `memory_/observe_/reflect_` → chronicler, unmatched → Unmatched bucket (not assigned to any agent).

#### Scenario: Operator gets shell, file, and skill tools
- **WHEN** tools named `exec_shell`, `fs_read`, `skill_deploy` are registered
- **THEN** they SHALL be assigned to the operator sub-agent

#### Scenario: Navigator gets browser tools
- **WHEN** tools named `browser_navigate`, `browser_screenshot` are registered
- **THEN** they SHALL be assigned to the navigator sub-agent

#### Scenario: Vault gets crypto, secrets, and payment tools
- **WHEN** tools named `crypto_sign`, `secrets_get`, `payment_send` are registered
- **THEN** they SHALL be assigned to the vault sub-agent

#### Scenario: Librarian gets search, RAG, graph, and skill management tools
- **WHEN** tools named `search_web`, `rag_query`, `graph_traverse`, `save_knowledge_item`, `create_skill_x`, `list_skills` are registered
- **THEN** they SHALL be assigned to the librarian sub-agent

#### Scenario: Chronicler gets memory tools
- **WHEN** tools named `memory_store`, `observe_event`, `reflect_summary` are registered
- **THEN** they SHALL be assigned to the chronicler sub-agent

#### Scenario: Unmatched tools tracked separately
- **WHEN** a tool with an unrecognized prefix is present
- **THEN** it SHALL be placed in the Unmatched bucket and NOT assigned to any sub-agent

#### Scenario: Librarian prefix priority over operator
- **WHEN** tools like `save_knowledge_data` or `create_skill_new` are registered
- **THEN** they SHALL match librarian prefixes before reaching operator matching

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
Sub-agent descriptions in the orchestrator prompt SHALL use human-readable capability summaries instead of raw tool names. The `capabilityMap` SHALL include entries for all tool prefixes including `secrets_`, `create_skill`, and `list_skills`. The `capabilityDescription()` function SHALL deduplicate capabilities across a tool set.

#### Scenario: Operator description uses capabilities
- **WHEN** the operator sub-agent has tools `exec_shell`, `fs_read`
- **THEN** its description SHALL contain "command execution, file operations"
- **AND** it SHALL NOT contain raw tool names

#### Scenario: Vault description uses capabilities
- **WHEN** the vault sub-agent has tools `crypto_sign`, `secrets_get`, `payment_send`
- **THEN** its description SHALL contain "cryptography, secret management, blockchain payments (USDC on Base)"

#### Scenario: Duplicate capabilities are deduplicated
- **WHEN** two tools share the same prefix (e.g., `exec_shell` and `exec_run`)
- **THEN** the capability "command execution" SHALL appear only once in the description

#### Scenario: Unknown tool prefix falls back to general actions
- **WHEN** a tool has no matching prefix in `capabilityMap`
- **THEN** its capability SHALL be "general actions"

### Requirement: Orchestrator instruction guides delegation-only execution
The orchestrator instruction SHALL enforce mandatory delegation for all tool-requiring tasks. It SHALL include a routing table with exact agent names, a decision protocol, and rejection handling. Sub-agent entries SHALL use capability descriptions, not raw tool name lists. The instruction SHALL NOT contain words that could be confused with agent names.

#### Scenario: Tool-requiring task
- **WHEN** a user requests any task requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent using its exact registered name

#### Scenario: Agent name exactness
- **WHEN** the orchestrator delegates to a sub-agent
- **THEN** it SHALL use the EXACT name (e.g. "operator", NOT "exec", "browser", or any abbreviation)

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
The `BuildAgentTree` function SHALL create sub-agents data-driven from the agentSpecs registry. Agents with no tools SHALL be skipped unless AlwaysInclude is set. The planner sub-agent SHALL always be created as it is LLM-only.

#### Scenario: All tool categories have tools
- **WHEN** tools exist for operator, navigator, vault, librarian, and chronicler roles
- **THEN** all six sub-agents (operator, navigator, vault, librarian, planner, chronicler) SHALL be created

#### Scenario: Partial tools — only operator and librarian
- **WHEN** only operator and librarian tools are provided
- **THEN** only operator, librarian, and planner sub-agents SHALL be created

#### Scenario: No tools at all
- **WHEN** the tool list is empty
- **THEN** only the planner sub-agent SHALL be created

#### Scenario: Unmatched-only tools
- **WHEN** all tools have unrecognized prefixes
- **THEN** only the planner sub-agent SHALL be created
- **AND** no unmatched tools SHALL be adapted

### Requirement: Orchestrator Short-Circuit for Simple Queries
The orchestrator instruction SHALL direct the LLM to respond directly to simple conversational queries (greetings, opinions, general knowledge) without delegating to sub-agents.

#### Scenario: Simple greeting
- **WHEN** user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Task requiring tools
- **WHEN** user requests an action requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

### Requirement: SubAgentPromptFunc type
The orchestration package SHALL define a `SubAgentPromptFunc` function type that takes `(agentName, defaultInstruction string)` and returns the assembled system prompt string for a sub-agent.

#### Scenario: Function receives correct parameters
- **WHEN** `BuildAgentTree` calls the `SubAgentPromptFunc` for each sub-agent
- **THEN** it SHALL pass the agent's spec name and the original spec.Instruction

### Requirement: Config supports SubAgentPrompt field
The orchestration `Config` struct SHALL include a `SubAgentPrompt SubAgentPromptFunc` field. When set, `BuildAgentTree` SHALL use it to build each sub-agent's instruction. When nil, the original `spec.Instruction` is used.

#### Scenario: SubAgentPrompt set
- **WHEN** `Config.SubAgentPrompt` is non-nil
- **THEN** `BuildAgentTree` SHALL call it for every sub-agent and use the returned string as the agent's Instruction

#### Scenario: SubAgentPrompt nil (backward compatible)
- **WHEN** `Config.SubAgentPrompt` is nil
- **THEN** `BuildAgentTree` SHALL use `spec.Instruction` directly, preserving existing behavior

### Requirement: Max Delegation Rounds
The `Config` struct SHALL include a `MaxDelegationRounds` field. The orchestrator instruction SHALL mention this limit as a prompt-based guardrail.

#### Scenario: Default max rounds
- **WHEN** `MaxDelegationRounds` is zero or unset
- **THEN** the default limit of 5 rounds SHALL be used in the orchestrator prompt

### Requirement: Dynamic Orchestrator Instruction
The orchestrator instruction SHALL be dynamically generated to list only the sub-agents that were actually created, rather than hardcoding all agent names.

#### Scenario: Only operator and planner created
- **WHEN** only operator and planner sub-agents are created
- **THEN** the orchestrator instruction SHALL only mention operator and planner

### Requirement: Sub-Agent Result Reporting
Each sub-agent instruction SHALL include guidance to report results clearly after completing their task, structured with What You Do, Input Format, Output Format, and Constraints sections.

#### Scenario: Operator result reporting
- **WHEN** the operator sub-agent completes an action
- **THEN** its instruction SHALL guide it to provide results clearly

#### Scenario: Librarian result reporting
- **WHEN** the librarian sub-agent completes research
- **THEN** its instruction SHALL guide it to organize results clearly

#### Scenario: Planner result reporting
- **WHEN** the planner sub-agent completes planning
- **THEN** its instruction SHALL guide it to present the plan for review

#### Scenario: Chronicler result reporting
- **WHEN** the chronicler sub-agent completes memory operations
- **THEN** its instruction SHALL guide it to report what was stored or retrieved

### Requirement: RoleToolSet has six roles plus Unmatched
The RoleToolSet struct SHALL have fields: Operator, Navigator, Vault, Librarian, Planner, Chronicler, and Unmatched. Each field is a slice of `*agent.Tool`.

#### Scenario: RoleToolSet structure
- **WHEN** PartitionTools is called
- **THEN** it SHALL return a RoleToolSet with seven fields (six roles + Unmatched)

#### Scenario: Planner tools always empty
- **WHEN** PartitionTools is called with any input
- **THEN** the Planner field SHALL always be nil/empty
