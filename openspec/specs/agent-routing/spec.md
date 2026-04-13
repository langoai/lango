## Purpose

Multi-agent orchestration tool routing — prefix-based tool partitioning, agent spec registry, and orchestrator delegation protocol.

## Requirements

### Requirement: AgentSpec registry defines sub-agent identity and routing metadata
The system SHALL define an `AgentSpec` type with fields: Name, Description, Instruction, Prefixes, Keywords, Accepts, Returns, CannotDo, and AlwaysInclude. A `var agentSpecs` registry SHALL contain specs for all 6 sub-agents in creation order.

#### Scenario: AgentSpec type has all required fields
- **WHEN** the AgentSpec type is defined
- **THEN** it SHALL include Name (string), Description (string), Instruction (string), Prefixes ([]string), Keywords ([]string), Accepts (string), Returns (string), CannotDo ([]string), and AlwaysInclude (bool)

#### Scenario: Registry contains exactly 6 specs
- **WHEN** agentSpecs is initialized
- **THEN** it SHALL contain specs for operator, navigator, vault, librarian, planner, and chronicler in that order

#### Scenario: Each spec has unique name
- **WHEN** agentSpecs is iterated
- **THEN** no two specs SHALL have the same Name

#### Scenario: Librarian prefixes include all skill management tools
- **WHEN** the librarian spec's Prefixes are checked
- **THEN** they SHALL include `create_skill`, `list_skills`, and `import_skill`

#### Scenario: import_skill routes to librarian
- **WHEN** a tool named `import_skill` is partitioned via `PartitionTools`
- **THEN** it SHALL be assigned to the Librarian role tool set

#### Scenario: import_skill has capability description
- **WHEN** `toolCapability` is called with a tool name starting with `import_skill`
- **THEN** it SHALL return a non-empty capability description

### Requirement: Routing table in orchestrator prompt
The orchestrator instruction SHALL contain a structured routing table listing each sub-agent with its role, keywords, accepts/returns format, and cannot-do constraints.

#### Scenario: Routing table format
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL contain a `### <agent-name>` section for each active sub-agent
- **AND** each section SHALL include Role, Keywords, Accepts, Returns, and Cannot fields

#### Scenario: Only active agents in routing table
- **WHEN** some agents have no tools and are not AlwaysInclude
- **THEN** the routing table SHALL only list agents that were actually created

### Requirement: Decision protocol in orchestrator prompt
The orchestrator instruction SHALL include a 5-step decision protocol: CLASSIFY, MATCH, SELECT, VERIFY, DELEGATE.

#### Scenario: Decision protocol presence
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL contain the steps CLASSIFY, MATCH, SELECT, VERIFY, and DELEGATE

### Requirement: Reject protocol for misrouted tasks
Each sub-agent instruction SHALL include a `[REJECT]` response protocol. When a sub-agent receives a task outside its capabilities, it SHALL respond with `[REJECT] This task requires <correct_agent>. I handle: <capabilities>.`

#### Scenario: All sub-agent instructions contain reject protocol
- **WHEN** any sub-agent's instruction is checked
- **THEN** it SHALL contain the string `[REJECT]`

#### Scenario: Orchestrator handles rejections
- **WHEN** a sub-agent rejects a task
- **THEN** the orchestrator SHALL try the next most relevant agent or handle directly

### Requirement: Keywords for routing decisions
Each AgentSpec SHALL define keywords that the orchestrator uses to match user requests to agents.

#### Scenario: All specs have keywords
- **WHEN** agentSpecs is iterated
- **THEN** every spec SHALL have a non-empty Keywords slice

#### Scenario: Operator keywords
- **WHEN** the operator spec is checked
- **THEN** its Keywords SHALL include: run, execute, command, shell, file

#### Scenario: Navigator keywords
- **WHEN** the navigator spec is checked
- **THEN** its Keywords SHALL include: browse, web, url, page, navigate

#### Scenario: Vault keywords
- **WHEN** the vault spec is checked
- **THEN** its Keywords SHALL include: encrypt, decrypt, sign, secret, payment, wallet

### Requirement: Unmatched tools reported in orchestrator prompt
When tools match no agent prefix, the orchestrator instruction SHALL list them under an "Unmatched Tools" section.

#### Scenario: Unmatched tools present
- **WHEN** unmatched tools exist
- **THEN** the orchestrator instruction SHALL contain "Unmatched Tools" with the tool names listed

#### Scenario: No unmatched tools
- **WHEN** all tools match a prefix
- **THEN** the orchestrator instruction SHALL NOT contain "Unmatched Tools"

### Requirement: Sub-agent instructions follow structured format
Every sub-agent instruction SHALL contain four sections: "What You Do", "Input Format", "Output Format", and "Constraints".

#### Scenario: Instruction structure
- **WHEN** any agentSpec's instruction is checked
- **THEN** it SHALL contain `## What You Do`, `## Input Format`, `## Output Format`, and `## Constraints`

### Requirement: Accepts and Returns metadata
Every AgentSpec SHALL define Accepts (input format description) and Returns (output format description) for the routing table.

#### Scenario: All specs have I/O metadata
- **WHEN** agentSpecs is iterated
- **THEN** every spec SHALL have non-empty Accepts and Returns fields

### Requirement: Vault agent prefix list covers all vault tool families
The vault AgentSpec Prefixes SHALL include prefixes for all vault-domain tool families: crypto, secrets, payment, p2p, smartaccount, economy, escrow, sentinel, and contract.

#### Scenario: Vault prefixes include smartaccount tools
- **WHEN** the vault spec's Prefixes are checked
- **THEN** they SHALL include `smart_account_`, `session_key_`, `session_execute`, `policy_check`, `module_`, `spending_`, `paymaster_`

#### Scenario: Vault prefixes include economy/escrow/sentinel/contract tools
- **WHEN** the vault spec's Prefixes are checked
- **THEN** they SHALL include `economy_`, `escrow_`, `sentinel_`, `contract_`

### Requirement: Dynamic and builtin vault specs are synchronized
The embedded AGENT.md for vault and the builtin vault spec in agentSpecs SHALL have identical prefix and keyword lists.

#### Scenario: AGENT.md prefixes match builtin spec
- **WHEN** the vault AGENT.md frontmatter prefixes are compared to agentSpecs vault Prefixes
- **THEN** the lists SHALL contain the same entries

### Requirement: capabilityMap covers all vault prefixes
Every prefix in the vault AgentSpec SHALL have a corresponding entry in capabilityMap for diagnostics.

#### Scenario: All vault prefixes have capability entries
- **WHEN** toolCapability is called with tool names starting with any vault prefix
- **THEN** it SHALL return a non-empty capability description

### Requirement: AgentSpec routing metadata
AgentSpec SHALL include ExampleRequests and Disambiguation fields for precise routing.

#### Scenario: ExampleRequests field
- **WHEN** an AgentSpec is defined
- **THEN** it SHALL have an ExampleRequests field containing 3-5 concrete request examples

#### Scenario: Disambiguation field
- **WHEN** an AgentSpec is defined
- **THEN** it SHALL have a Disambiguation field explaining when NOT to pick this agent

#### Scenario: Routing entry propagation
- **WHEN** a routingEntry is built from an AgentSpec
- **THEN** it SHALL propagate ExampleRequests and Disambiguation from the spec

### Requirement: Keywords use compound phrases
Agent keywords SHALL use compound phrases instead of ambiguous single words to reduce routing overlap.

#### Scenario: Operator keywords
- **WHEN** the operator agent keywords are checked
- **THEN** they SHALL contain "run command", "execute command", "terminal" instead of bare "run", "execute"

#### Scenario: Librarian keywords
- **WHEN** the librarian agent keywords are checked
- **THEN** they SHALL contain "search knowledge", "find information", "save knowledge" instead of bare "search", "find"

#### Scenario: Chronicler keywords
- **WHEN** the chronicler agent keywords are checked
- **THEN** they SHALL contain "remember this", "recall conversation", "conversation memory" instead of bare "remember", "memory"

### Requirement: Universal tool_output_ distribution
Tools with "tool_output_" prefix SHALL be distributed to all non-empty, tool-bearing agent sets.

#### Scenario: Distribution to active agents
- **WHEN** PartitionTools is called with a tool_output_get tool and multiple agents have tools
- **THEN** tool_output_get SHALL appear in every non-empty agent tool set

#### Scenario: No distribution to empty agents
- **WHEN** PartitionTools is called and an agent has no tools
- **THEN** tool_output_get SHALL NOT be added to that agent's empty tool set

#### Scenario: Planner exclusion
- **WHEN** PartitionTools distributes universal tools
- **THEN** planner SHALL NOT receive tool_output_get (it has no tools)

#### Scenario: No duplicate distribution
- **WHEN** tool_output_get is distributed to an agent
- **THEN** it SHALL appear exactly once in that agent's tool set

#### Scenario: Dynamic partition support
- **WHEN** PartitionToolsDynamic is called with tool_output_ tools
- **THEN** the same universal distribution logic SHALL apply

### Requirement: Prompt override file consistency
All prompt override files (IDENTITY.md, AGENT.md) SHALL use transfer_to_agent escalation instead of [REJECT] patterns.

#### Scenario: No REJECT patterns
- **WHEN** any prompt override file is checked
- **THEN** it SHALL NOT contain the text "[REJECT]"

#### Scenario: Escalation protocol present
- **WHEN** any prompt override file is checked
- **THEN** it SHALL contain "transfer_to_agent" escalation to "lango-orchestrator"

#### Scenario: Output handling in non-planner overrides
- **WHEN** a non-planner prompt override file is checked
- **THEN** it SHALL contain "## Output Handling" section with tool_output_get guidance

