## MODIFIED Requirements

### Requirement: Agent memory store persistence
The agent memory system SHALL use a persistent Ent-backed store (`EntStore`) as the default backend instead of the in-memory store. Agent memories SHALL be retained across server restarts.

#### Scenario: Memory persists across restart
- **WHEN** an agent saves a memory entry and the server is restarted
- **THEN** the memory entry SHALL be retrievable after restart

#### Scenario: InMemoryStore retained for testing
- **WHEN** tests require an agent memory store
- **THEN** `InMemoryStore` SHALL remain available as an alternative backend

### Requirement: Per-agent memory isolation via name propagation
Each agent's memory entries SHALL be isolated by agent name. The agent name SHALL be propagated through tool context for both root agent and sub-agent paths.

#### Scenario: Root agent stores under "lango-agent"
- **WHEN** the root agent calls `memory_agent_save`
- **THEN** the entry SHALL be stored with `agent_name = "lango-agent"`

#### Scenario: Sub-agent stores under its own name
- **WHEN** a sub-agent named "researcher" calls `memory_agent_save`
- **THEN** the entry SHALL be stored with `agent_name = "researcher"`

#### Scenario: Agents cannot see each other's instance-scoped entries
- **WHEN** agent "researcher" saves an entry with scope "instance"
- **THEN** `SearchWithContext("executor", ...)` SHALL NOT return that entry
