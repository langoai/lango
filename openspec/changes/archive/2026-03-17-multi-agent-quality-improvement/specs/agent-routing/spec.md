## MODIFIED Requirements

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

## ADDED Requirements

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
