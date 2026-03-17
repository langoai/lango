## MODIFIED Requirements

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

### Requirement: Round budget guidance
The round budget guidance SHALL prioritize task completion over premature summarization.

#### Scenario: Low rounds guidance
- **WHEN** the orchestrator is running low on delegation rounds
- **THEN** it SHALL prioritize completing the current step and transparently report what remains if unable to finish

### Requirement: Output awareness
The orchestrator instruction SHALL include output awareness guidance for compressed tool output.

#### Scenario: Output awareness section present
- **WHEN** the orchestrator instruction is generated
- **THEN** it SHALL contain an "Output Awareness" section describing _meta.compressed handling

## ADDED Requirements

### Requirement: Sub-agent output handling section
All non-planner sub-agent instructions SHALL include an Output Handling section teaching agents to use `tool_output_get` for compressed results.

#### Scenario: Non-planner agents have output handling
- **WHEN** a sub-agent instruction is generated for operator, navigator, vault, librarian, automator, or chronicler
- **THEN** the instruction SHALL contain "## Output Handling" with `tool_output_get` guidance

#### Scenario: Planner excluded from output handling
- **WHEN** a sub-agent instruction is generated for planner
- **THEN** the instruction SHALL NOT contain "## Output Handling"
