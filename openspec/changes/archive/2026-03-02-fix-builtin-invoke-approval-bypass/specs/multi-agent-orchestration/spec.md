## MODIFIED Requirements

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
