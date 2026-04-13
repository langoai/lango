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

#### Scenario: Navigator bounded search protocol
- **WHEN** the navigator handles a topic-based live web request
- **THEN** its instruction SHALL direct it to run `browser_search` once and then prefer current-page extraction over repeated search
- **AND** it SHALL allow at most one search reformulation when the first results are empty or clearly unrelated
- **AND** it SHALL stop once the requested count of credible results has been collected
