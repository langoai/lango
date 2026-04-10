## MODIFIED Requirements

### Requirement: Hierarchical agent tree with sub-agents
The system SHALL support a multi-agent mode (`agent.multiAgent: true`) that creates an orchestrator root agent with specialized sub-agents: operator, navigator, vault, librarian, automator, planner, and chronicler. The orchestrator SHALL have NO direct tools (`Tools: nil`) and MUST delegate all tool-requiring tasks to sub-agents. Each sub-agent SHALL include an Escalation Protocol section in its instruction that directs it to call `transfer_to_agent` with agent_name `lango-orchestrator` when it receives an out-of-scope request. Sub-agents SHALL NOT emit `[REJECT]` text or tell users to ask another agent.

#### Scenario: Multi-agent mode enabled
- **WHEN** `agent.multiAgent` is true
- **THEN** BuildAgentTree SHALL create an orchestrator that has NO direct tools AND has sub-agents (operator, navigator, vault, librarian, automator, planner, chronicler)

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
- **WHEN** agentSpecs are defined for all 7 sub-agents
- **THEN** every spec's Instruction SHALL contain `transfer_to_agent` and `lango-orchestrator`
- **AND** every spec's Instruction SHALL contain `## Escalation Protocol`

## ADDED Requirements

### Requirement: Orchestrator direct response assessment
The orchestrator's Decision Protocol SHALL include a Step 0 (ASSESS) that evaluates whether a request is a simple conversational message (greeting, general knowledge, opinion, weather, math, small talk). If yes, the orchestrator SHALL respond directly without delegation.

#### Scenario: Simple greeting handled directly
- **WHEN** the user sends a greeting like "Hello"
- **THEN** the orchestrator SHALL respond directly without delegating to any sub-agent

#### Scenario: General knowledge handled directly
- **WHEN** the user asks a general knowledge question (e.g., weather, math)
- **THEN** the orchestrator SHALL respond directly without delegation

#### Scenario: Tool-requiring request delegated normally
- **WHEN** the user requests an action requiring tools (e.g., "create a wallet")
- **THEN** the orchestrator SHALL delegate to the appropriate sub-agent per the routing table

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
