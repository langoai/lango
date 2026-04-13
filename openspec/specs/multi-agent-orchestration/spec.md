## Purpose

Hierarchical multi-agent orchestration for Lango. Defines how the tool-less orchestrator routes work to specialist sub-agents, how tools are partitioned, and what runtime delegation guarantees the system must uphold.

## Requirements

### Requirement: Orchestrator universal tools
The orchestration `Config` struct SHALL include a `UniversalTools` field. In multi-agent mode, the orchestrator SHALL NOT receive universal tools. `BuildAgentTree` SHALL NOT adapt or assign `UniversalTools` to the orchestrator agent. The orchestrator SHALL have no direct tools and MUST delegate all tasks to sub-agents.

#### Scenario: Multi-agent orchestrator has no tools
- **WHEN** `BuildAgentTree` is called (multi-agent mode)
- **THEN** the orchestrator agent SHALL have no tools (Tools is nil/empty)
- **AND** the orchestrator instruction SHALL state "You do NOT have tools"
- **AND** the instruction SHALL NOT mention builtin_list or builtin_invoke

#### Scenario: Config.UniversalTools field preserved
- **WHEN** `Config.UniversalTools` is set
- **THEN** the field SHALL be accepted without error but SHALL NOT be wired to the orchestrator

### Requirement: Builtin prefix exclusion from partitioning
`PartitionTools` SHALL skip any tool whose name starts with `builtin_`. These tools SHALL NOT appear in any sub-agent's tool set or in the Unmatched list.

#### Scenario: Builtin tools skipped during partitioning
- **WHEN** tools include `builtin_list` and `builtin_invoke` alongside normal tools
- **THEN** `PartitionTools` SHALL assign normal tools to their respective roles
- **AND** `builtin_*` tools SHALL not appear in any RoleToolSet field

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with specialized sub-agents: operator, navigator, vault, librarian, automator, planner, chronicler, and ontologist. The orchestrator SHALL have NO direct tools (`Tools: nil`) and MUST delegate all tool-requiring tasks to sub-agents. Each sub-agent SHALL include an Escalation Protocol section in its instruction that directs it to call `transfer_to_agent` with agent_name `lango-orchestrator` when it receives an out-of-scope request. Sub-agents SHALL NOT emit `[REJECT]` text or tell users to ask another agent.

#### Scenario: Multi-agent mode enabled
- **WHEN** `agent.multiAgent` is true
- **THEN** BuildAgentTree SHALL create an orchestrator that has NO direct tools AND has sub-agents (operator, navigator, vault, librarian, automator, planner, chronicler, ontologist)

#### Scenario: Orchestrator has no direct tools
- **WHEN** the orchestrator is created
- **THEN** the orchestrator's `Tools` field SHALL be `nil`
- **AND** tools SHALL only be adapted for their respective sub-agents (each tool adapted exactly once)

#### Scenario: Single-agent fallback
- **WHEN** `agent.multiAgent` is false
- **THEN** the system SHALL create a single flat agent with all tools

#### Scenario: Sub-agent escalation via transfer_to_agent
- **WHEN** a sub-agent receives a request outside its capabilities
- **THEN** the sub-agent instruction SHALL direct it to call `transfer_to_agent` with agent_name `lango-orchestrator`
- **AND** the sub-agent SHALL NOT emit any text before the transfer call
- **AND** the sub-agent instruction SHALL contain `## Escalation Protocol` section

#### Scenario: All sub-agents have escalation protocol
- **WHEN** agentSpecs are defined for all 8 sub-agents
- **THEN** every spec's Instruction SHALL contain `transfer_to_agent` and `lango-orchestrator`
- **AND** every spec's Instruction SHALL contain `## Escalation Protocol`

#### Scenario: Navigator fallback protocol
- **WHEN** the navigator receives a live web query and `browser_search` is unavailable in the current runtime
- **THEN** its instruction SHALL direct it to continue with `browser_navigate` to a search URL and `browser_extract` in `search_results` mode
- **AND** if those higher-level tools are also unavailable, it SHALL continue with low-level `browser_action` or `eval` rather than stopping while browser browsing remains in scope

#### Scenario: Navigator bounded search protocol
- **WHEN** the navigator handles a topic-based live web request
- **THEN** its instruction SHALL direct it to run `browser_search` once and then prefer current-page extraction over repeated search
- **AND** it SHALL allow at most one search reformulation when the first results are empty or clearly unrelated
- **AND** it SHALL stop once the requested count of credible results has been collected

### Requirement: Tool partitioning by prefix
Tools SHALL be partitioned to sub-agents based on name prefixes with matching order Librarian → Chronicler → Ontologist → Navigator → Vault → Operator → Unmatched: `exec/fs_/skill_` → operator, `browser_` → navigator, `crypto_/secrets_/payment_` → vault, `search_/rag_/graph_/save_knowledge/save_learning/create_skill/list_skills` → librarian, `memory_/observe_/reflect_` → chronicler, `ontology_` → ontologist, unmatched → Unmatched bucket (not assigned to any agent).

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

#### Scenario: Capability description includes librarian inquiry tools
- **WHEN** capabilityDescription is called for a tool set containing `librarian_pending_inquiries`
- **THEN** the description includes "knowledge inquiries and gap detection"

### Requirement: Orchestrator instruction guides delegation-only execution
The orchestrator instruction SHALL enforce mandatory delegation for all tool-requiring tasks. It SHALL include a routing table with exact agent names, a decision protocol, and rejection handling. Sub-agent entries SHALL use capability descriptions, not raw tool name lists. The instruction SHALL NOT contain words that could be confused with agent names. The instruction SHALL always state the orchestrator has no tools.

#### Scenario: Tool-requiring task
- **WHEN** a user requests any task requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent using its exact registered name

#### Scenario: Delegation-only prompt
- **WHEN** the orchestrator instruction is built
- **THEN** it SHALL contain "You do NOT have tools"
- **AND** it SHALL contain "MUST delegate all tool-requiring tasks"
- **AND** it SHALL NOT contain "builtin_list" or "builtin_invoke"

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
- **WHEN** tools exist for operator, navigator, vault, librarian, automator, and chronicler roles
- **THEN** all eight sub-agents (operator, navigator, vault, librarian, automator, planner, chronicler, ontologist) SHALL be created

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

### Requirement: Orchestrator direct response assessment
The orchestrator's Decision Protocol SHALL include a Step 0 (ASSESS) that evaluates whether a request is a simple conversational message (greeting, opinion, math, small talk). If yes, the orchestrator SHALL respond directly without delegation. The ASSESS step SHALL explicitly state that the orchestrator MUST NOT emit any function calls even when responding directly, and that requests requiring real-time data (weather, news, prices, search) SHALL be delegated to the appropriate sub-agent.

#### Scenario: Simple greeting handled directly
- **WHEN** the user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegating to any sub-agent

#### Scenario: Math question handled directly
- **WHEN** the user asks a math question (e.g., "what is 2 + 2?")
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Weather request delegated
- **WHEN** the user asks about weather (e.g., "tell me today's weather")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent (e.g., navigator or librarian)
- **AND** the orchestrator SHALL NOT attempt to respond directly or emit any function calls

#### Scenario: General knowledge requiring search delegated
- **WHEN** the user asks a factual question that may require search (e.g., "what is the latest news?")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

#### Scenario: Tool-requiring request delegated normally
- **WHEN** the user requests an action requiring tools (e.g., "create a wallet")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent per the routing table

#### Scenario: No function calls in direct response
- **WHEN** the orchestrator responds directly (ASSESS step determines direct-answer)
- **THEN** the orchestrator SHALL NOT emit any function calls
- **AND** the ASSESS instruction SHALL contain "MUST NOT emit any function calls"

### Requirement: Orchestrator re-routing protocol
The orchestrator instruction SHALL include a "Re-Routing Protocol" section. When a sub-agent transfers control back to the orchestrator, the orchestrator SHALL NOT re-send the same request to the same agent. It SHALL re-evaluate using the Decision Protocol (starting from Step 0) and either route to a different agent or answer directly as a general-purpose assistant.

#### Scenario: Sub-agent transfers back
- **WHEN** a sub-agent calls `transfer_to_agent` to return control to the orchestrator
- **THEN** the orchestrator SHALL re-evaluate the request using the Decision Protocol from Step 0
- **AND** SHALL NOT re-send to the same sub-agent

#### Scenario: No matching agent after re-evaluation
- **WHEN** re-evaluation determines no sub-agent can handle the request
- **THEN** the orchestrator SHALL answer the question itself as a general-purpose assistant

### Requirement: Delegation rules prioritize direct response
The orchestrator's Delegation Rules SHALL list direct response for simple conversational messages BEFORE the rule about delegating tool-requiring tasks.

#### Scenario: Delegation rules ordering
- **WHEN** the orchestrator instruction is built
- **THEN** the rule about responding directly to simple messages SHALL appear before the rule about delegating to sub-agents

### Requirement: Orchestrator Short-Circuit for Simple Queries
The orchestrator instruction SHALL direct the LLM to respond directly to simple conversational queries (greetings, opinions, math, small talk) without delegating to sub-agents.

#### Scenario: Simple greeting
- **WHEN** user sends a greeting like "hello"
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Task requiring tools
- **WHEN** user requests an action requiring tool execution
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent

#### Scenario: Direct-answer list excludes real-time topics
- **WHEN** the orchestrator instruction is built
- **THEN** the direct-answer categories in ASSESS and Delegation Rules SHALL NOT include "weather" or "general knowledge"

### Requirement: SubAgentPromptFunc type
The orchestration package SHALL expose a `SubAgentPromptFunc` type that can rewrite default sub-agent instructions at build time.

#### Scenario: Prompt override hook available
- **WHEN** multi-agent orchestration is initialized
- **THEN** callers SHALL be able to provide a `SubAgentPromptFunc`
- **AND** the returned string SHALL replace the default sub-agent instruction for that agent

### Requirement: Specialist completion contract
After a specialist sub-agent receives successful tool output, the runtime SHALL require the specialist turn to end in exactly one of: visible assistant completion, `transfer_to_agent`, or a structured incomplete outcome emitted by the runtime. Tool-only terminal states SHALL NOT be treated as successful completion.

#### Scenario: Successful tool use followed by visible completion
- **WHEN** `vault` successfully retrieves wallet balance information
- **THEN** the specialist turn SHALL produce a visible assistant completion summarizing the result
- **AND** the runtime SHALL classify the turn as successful

#### Scenario: Tool-only terminal state becomes incomplete outcome
- **WHEN** a specialist receives successful tool output but terminates without visible completion or transfer
- **THEN** the runtime SHALL terminate the specialist turn with a structured incomplete outcome
- **AND** the parent turn SHALL NOT treat the specialist turn as a silent success

### Requirement: Repeated identical specialist call containment
Within a single user turn, repeated calls from the same specialist to the same tool with canonically equal params SHALL be detected and stopped even if the model changes call IDs between attempts.

#### Scenario: Same tool same params repeated with different call IDs
- **WHEN** `vault` repeatedly calls `payment_balance` with `{}` and each attempt has a different call ID
- **THEN** the runtime SHALL still count those attempts against the same call-signature loop budget
- **AND** SHALL stop the loop when the threshold is reached

#### Scenario: Different params do not trip the identical-call budget immediately
- **WHEN** the same specialist calls the same tool name with materially different params
- **THEN** the runtime SHALL treat those calls as distinct signatures for loop containment purposes

### Requirement: Evidence-only orchestrator recovery
When the orchestrator has no direct tools and a delegated specialist fails or returns an incomplete outcome, the orchestrator SHALL either re-route to another agent or answer only from evidence already gathered in the turn trace. It SHALL NOT emit direct FunctionCalls to specialist-only tools.

#### Scenario: Tool-less orchestrator cannot call specialist tool directly
- **WHEN** a previous `vault` turn failed and the orchestrator enters recovery
- **THEN** the orchestrator SHALL NOT emit a direct FunctionCall to `payment_balance`
- **AND** SHALL instead re-route, answer from existing evidence, or report an inability to complete

#### Scenario: Recovery answer references gathered evidence only
- **WHEN** the orchestrator answers after a specialist failure
- **THEN** the answer SHALL be derived only from tool results or summaries already recorded in the current turn trace
- **AND** it SHALL NOT claim that a new unavailable tool call was executed
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
- **THEN** the default limit of 10 rounds SHALL be used in the orchestrator prompt

### Requirement: Round budget guidance
The round budget guidance SHALL prioritize task completion over premature summarization.

#### Scenario: Low rounds guidance
- **WHEN** the orchestrator is running low on delegation rounds
- **THEN** it SHALL prioritize completing the current step and transparently report what remains if unable to finish

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

### Requirement: RoleToolSet has eight roles plus Unmatched
The RoleToolSet struct SHALL have fields: Operator, Navigator, Vault, Librarian, Planner, Chronicler, Automator, Ontologist, and Unmatched. Each field is a slice of `*agent.Tool`.

#### Scenario: RoleToolSet structure
- **WHEN** PartitionTools is called
- **THEN** it SHALL return a RoleToolSet with nine fields (eight roles + Unmatched)

#### Scenario: Planner tools always empty
- **WHEN** PartitionTools is called with any input
- **THEN** the Planner field SHALL always be nil/empty

### Requirement: Librarian Agent Specification
The librarian sub-agent SHALL handle knowledge management including: search, RAG, graph traversal, knowledge/skill persistence, and knowledge inquiries. The agent spec SHALL include `librarian_` in its Prefixes list and `inquiry`, `question`, `gap` in its Keywords list. The Instruction SHALL include a "Proactive Behavior" section instructing the agent to weave pending inquiries naturally into responses.

#### Scenario: Librarian tool routing with inquiry prefix
- **WHEN** a tool named `librarian_pending_inquiries` is partitioned
- **THEN** it is assigned to the librarian sub-agent's tool set

#### Scenario: Inquiry keyword routing
- **WHEN** the orchestrator receives a request containing "inquiry" or "gap"
- **THEN** the routing table matches the librarian agent via keyword matching

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

### Requirement: Multi-agent default turn limit
When `agent.multiAgent` is true and no explicit `MaxTurns` is configured, the system SHALL default to 75 turns instead of the standard 50. This provides sufficient headroom for multi-agent workflows with delegation overhead.

#### Scenario: Multi-agent mode with no explicit MaxTurns
- **WHEN** `agent.multiAgent` is true AND `agent.maxTurns` is zero or unset
- **THEN** the system SHALL use 75 as the maximum turn limit

#### Scenario: Multi-agent mode with explicit MaxTurns
- **WHEN** `agent.multiAgent` is true AND `agent.maxTurns` is set to a positive value
- **THEN** the system SHALL use the explicitly configured value, not the multi-agent default

#### Scenario: Single-agent mode unaffected
- **WHEN** `agent.multiAgent` is false
- **THEN** the system SHALL use the standard default of 50 turns

### Requirement: Dynamic specs support in Config
The orchestration `Config` struct SHALL include a `Specs []AgentSpec` field. When non-nil, `BuildAgentTree` SHALL use these specs instead of the hardcoded built-in specs.

#### Scenario: Custom specs provided
- **WHEN** Config.Specs is set to a non-nil slice of AgentSpec
- **THEN** BuildAgentTree SHALL use those specs for agent tree construction

#### Scenario: Nil specs falls back to builtins
- **WHEN** Config.Specs is nil
- **THEN** BuildAgentTree SHALL use the default BuiltinSpecs()

### Requirement: DynamicAgents provider in Config
The orchestration `Config` struct SHALL include a `DynamicAgents` field of type `agentpool.DynamicAgentProvider`. When set, dynamic P2P agents SHALL appear in the orchestrator's routing table.

#### Scenario: P2P agents in routing table
- **WHEN** DynamicAgents is set and has available agents
- **THEN** each P2P agent SHALL appear in the routing table with "p2p:" prefix, trust score, and capabilities

#### Scenario: No P2P agents
- **WHEN** DynamicAgents is nil
- **THEN** the routing table SHALL contain only local and A2A agents

### Requirement: Capability-enhanced routing entries
Routing table entries SHALL include a `Capabilities` field listing the agent's capabilities. The orchestrator instruction SHALL display capabilities alongside agent descriptions.

#### Scenario: Routing entry with capabilities
- **WHEN** a routing entry is generated for an agent with capabilities ["search", "rag"]
- **THEN** the entry SHALL include those capabilities in the orchestrator instruction

### Requirement: DynamicToolSet and PartitionToolsDynamic
The orchestration package SHALL provide `DynamicToolSet` (map[string][]*agent.Tool) and `PartitionToolsDynamic(tools, specs)` function. The existing `PartitionTools()` SHALL be preserved as a backward-compatible wrapper.

#### Scenario: Dynamic partitioning matches static
- **WHEN** PartitionToolsDynamic is called with the built-in specs
- **THEN** the result SHALL match PartitionTools for the same tool set

#### Scenario: PartitionTools still works
- **WHEN** PartitionTools is called
- **THEN** it SHALL return the same results as before (backward compatible)

### Requirement: Diagnostics section in orchestrator prompt
The orchestrator system prompt SHALL include a Diagnostics section instructing the orchestrator to use `builtin_list` or `builtin_health` when tools appear to be missing or a feature is not working.

#### Scenario: Orchestrator prompt contains diagnostics guidance
- **WHEN** `buildOrchestratorInstruction()` generates the orchestrator prompt
- **THEN** the prompt SHALL contain a "Diagnostics" section
- **AND** the section SHALL reference `builtin_health` as the diagnostic tool

### Requirement: Orchestrator prompt structure
The orchestrator instruction SHALL NOT contain contradictions between its tool-less role and action instructions. The prompt SHALL NOT instruct the orchestrator to "handle directly" for unmatched tools or reference diagnostic tools it cannot access.

#### Scenario: Unmatched tools routing
- **WHEN** tools exist that match no agent prefix
- **THEN** the orchestrator instruction SHALL state to route to the best-matching agent by role, or inform the user the capability is not available

#### Scenario: No diagnostics section
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL NOT contain a "Diagnostics" section or references to `builtin_list`/`builtin_health`

### Requirement: Orchestrator routing table format
The routing table SHALL display agent capabilities, example requests, and disambiguation hints as primary routing signals. Tool names SHALL NOT be listed individually.

#### Scenario: Tool count instead of names
- **WHEN** an agent has assigned tools
- **THEN** the routing table SHALL show "Tool count: N" instead of listing individual tool names

#### Scenario: Example requests displayed
- **WHEN** an agent has ExampleRequests defined
- **THEN** the routing table SHALL display them as a bulleted list under "Example Requests"

#### Scenario: Disambiguation displayed
- **WHEN** an agent has a Disambiguation string defined
- **THEN** the routing table SHALL display it under "When NOT this agent"

### Requirement: Disambiguation rules
The orchestrator instruction SHALL include explicit disambiguation rules for overlapping keywords.

#### Scenario: Search disambiguation
- **WHEN** the user says "search" without a URL
- **THEN** the disambiguation rules SHALL direct to librarian

#### Scenario: Memory disambiguation
- **WHEN** the user says "memory" in a conversation context
- **THEN** the disambiguation rules SHALL direct to chronicler

### Requirement: Complexity analysis phase
The decision protocol SHALL include a complexity analysis phase before routing.

#### Scenario: Simple task routing
- **WHEN** a task involves 1 domain
- **THEN** the orchestrator SHALL route directly without planner involvement

#### Scenario: Complex task decomposition
- **WHEN** a task involves 3+ domains
- **THEN** the orchestrator SHALL delegate to planner first

### Requirement: Re-routing protocol prevents loops
The re-routing protocol SHALL prevent delegation loops by tracking failed agents.

#### Scenario: Failed agent exclusion
- **WHEN** a sub-agent transfers control back to the orchestrator
- **THEN** the orchestrator SHALL NOT re-delegate to that agent for the same request

#### Scenario: Consecutive failure fallback
- **WHEN** two consecutive agents fail for the same request
- **THEN** the orchestrator SHALL answer directly as a general-purpose assistant

### Requirement: Output awareness
The orchestrator instruction SHALL include output awareness guidance for compressed tool output.

#### Scenario: Output awareness section present
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL contain an "Output Awareness" section describing _meta.compressed handling

### Requirement: Sub-agent output handling section
All non-planner sub-agent instructions SHALL include an Output Handling section teaching agents to use `tool_output_get` for compressed results.

#### Scenario: Non-planner agents have output handling
- **WHEN** a sub-agent instruction is generated for operator, navigator, vault, librarian, automator, chronicler, or ontologist
- **THEN** the instruction SHALL contain "## Output Handling" with `tool_output_get` guidance

#### Scenario: Planner excluded from output handling
- **WHEN** a sub-agent instruction is generated for planner
- **THEN** the instruction SHALL NOT contain "## Output Handling"

### Requirement: Session Isolation
`SessionIsolation` SHALL be a runtime behavior contract, not metadata only.

#### Scenario: Isolated sub-agent uses same-run overlay
- **WHEN** a sub-agent with `SessionIsolation=true` runs
- **THEN** its raw events are written to child session history
- **AND** the active parent session's in-memory view SHALL also see those events for the current run
- **AND** the parent persistent history SHALL not store those raw events

#### Scenario: Successful isolated run summary-merges to parent
- **WHEN** an isolated child run completes successfully
- **THEN** any raw in-memory overlay for that child SHALL be removed from the parent view
- **AND** the parent session receives only a summary message
- **AND** the full child history is not appended to the parent persistent history

#### Scenario: Failed isolated run leaves compact failure note
- **WHEN** an isolated child run fails or returns only a rejection/escalation path
- **THEN** the child session is discarded
- **AND** any raw in-memory overlay for that child SHALL be removed from the parent view
- **AND** the parent session SHALL retain only a compact root-authored failure note

#### Scenario: Non-isolated sub-agent unchanged
- **WHEN** a sub-agent has `SessionIsolation=false`
- **THEN** it continues to use the existing parent-session execution path

### Requirement: DelegationGuard monitors orchestrator delegations
The `DelegationGuard` SHALL observe delegation events emitted by the root orchestrator and maintain per-agent circuit breaker state. When a circuit-open agent is targeted, the guard SHALL log a warning and publish a `CircuitBreakerTrippedEvent`. The guard SHALL NOT block or redirect delegations — routing authority remains with the root orchestrator LLM.

#### Scenario: Warn on delegation to circuit-open agent
- **WHEN** root orchestrator delegates to an agent whose circuit is open
- **THEN** DelegationGuard SHALL log a warning with agent name and circuit state

### Requirement: Doctor multi-agent checks extended
The existing `MultiAgentCheck` in doctor SHALL be extended with:
- Loop frequency: count traces with `outcome=loop_detected` in last 24h via `RecentByOutcome`, warn if >3
- Timeout frequency: count traces with `outcome=timeout` in last 24h, warn if >5
- Trace store growth: `TraceCount()` vs configured maxTraces, warn if >80%
- Average turn duration: mean `ended_at - started_at` of recent successful traces, warn if >2min

#### Scenario: Loop frequency warning
- **WHEN** 5 traces with `outcome=loop_detected` exist in the last 24 hours
- **THEN** doctor SHALL emit a warning with loop count and recommendation

#### Scenario: Trace growth warning
- **WHEN** `TraceCount()` returns 8500 and maxTraces is 10000
- **THEN** doctor SHALL emit a warning that trace store is at 85% capacity

### Requirement: Gateway delegation WebSocket events
The gateway SHALL broadcast `agent.delegation` and `agent.budget_warning` WebSocket events when the corresponding TurnRunner callbacks fire.

#### Scenario: Delegation event broadcast
- **WHEN** `Request.OnDelegation` callback fires with from="orchestrator", to="operator"
- **THEN** gateway SHALL broadcast `agent.delegation` event to session clients

### Requirement: TurnRunner delegation and budget callbacks
`turnrunner.Request` SHALL support optional `OnDelegation func(from, to, reason string)` and `OnBudgetWarning func(used, max int)` callbacks. These callbacks SHALL be invoked by the turn runner when delegation events are detected in the trace recorder and when delegation count approaches the configured threshold.

#### Scenario: Callback fires on delegation
- **WHEN** the trace recorder observes a delegation event and `Request.OnDelegation` is non-nil
- **THEN** the callback SHALL be invoked with the source agent, target agent, and reason

#### Scenario: Nil callback is no-op
- **WHEN** `Request.OnDelegation` is nil and a delegation event occurs
- **THEN** no callback SHALL be invoked and execution SHALL continue normally

### Requirement: Structured recovery avoids repeating a failed specialist
When structured orchestration handles a specialist failure, the recovery layer SHALL use the observed specialist identity to avoid immediately repeating the same failed specialist path.

#### Scenario: Specialist tool failure reroutes away from failed specialist
- **WHEN** the orchestrator delegates to `vault`, the specialist attempt fails with a tool error, and structured recovery is enabled
- **THEN** the recovery layer SHALL retry with a reroute hint instead of a blind same-input retry
- **AND** the hint SHALL identify `vault` as the failed specialist
- **AND** the hint SHALL instruct the orchestrator to choose a different specialist or answer directly

#### Scenario: Pre-specialist failure keeps generic retry behavior
- **WHEN** structured recovery handles a retryable failure before any specialist delegation is observed
- **THEN** the recovery layer MAY retry the same input
- **AND** it SHALL not fabricate a failed specialist identity

### Requirement: Automator agent routing includes agent and task prefixes
The automator agent spec SHALL include `"agent_"` and `"task_"` in its Prefixes list alongside `"cron_"`, `"bg_"`, and `"workflow_"`. The `capabilityMap` SHALL include entries mapping `"agent_"` to "agent lifecycle management" and `"task_"` to "structured task management". Tools with these prefixes SHALL be routed to the automator sub-agent.

#### Scenario: Agent lifecycle tools routed to automator
- **WHEN** tools with `agent_` prefix (e.g., `agent_spawn`, `agent_status`) are registered
- **THEN** they SHALL be partitioned to the Automator role in `PartitionTools`

#### Scenario: Task management tools routed to automator
- **WHEN** tools with `task_` prefix (e.g., `task_create`, `task_list`) are registered
- **THEN** they SHALL be partitioned to the Automator role in `PartitionTools`

#### Scenario: Capability descriptions for agent and task tools
- **WHEN** `toolCapability` is called for an `agent_` prefixed tool
- **THEN** it SHALL return "agent lifecycle management"
- **AND** `toolCapability` for a `task_` prefixed tool SHALL return "structured task management"
