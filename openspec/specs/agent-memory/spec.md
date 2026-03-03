## ADDED Requirements

### Requirement: MemoryEntry type
The `agentmemory` package SHALL define a MemoryEntry struct with fields: ID, AgentName, Scope, Kind, Key, Content, Confidence, UseCount, Tags, CreatedAt, UpdatedAt.

#### Scenario: MemoryEntry fields
- **WHEN** a MemoryEntry is created
- **THEN** it SHALL have all required fields for agent-scoped memory storage

### Requirement: MemoryScope resolution
Memory lookups SHALL follow scope resolution order: instance (specific agent instance) > type (agent type) > global (all agents). Higher-priority scopes SHALL override lower ones.

#### Scenario: Instance scope takes priority
- **WHEN** a memory key exists at both instance and type scope
- **THEN** the instance-scope entry SHALL be returned

#### Scenario: Fallback to global scope
- **WHEN** a memory key exists only at global scope
- **THEN** the global-scope entry SHALL be returned

### Requirement: In-memory MemStore
The package SHALL provide a `MemStore` implementation using sync.RWMutex-protected maps. It SHALL support Save (upsert), Get, Search, Delete, IncrementUseCount, and Prune operations.

#### Scenario: Save upserts by key
- **WHEN** Save is called with an existing key
- **THEN** the entry SHALL be updated (not duplicated)

#### Scenario: Search by agent and tags
- **WHEN** Search is called with agent name and tags
- **THEN** it SHALL return all matching entries sorted by use count descending

### Requirement: Agent memory tools
The `app` package SHALL register three agent memory tools: `memory_agent_save`, `memory_agent_recall`, `memory_agent_forget`.

#### Scenario: Save tool stores memory
- **WHEN** `memory_agent_save` is called with key, content, and scope
- **THEN** the entry SHALL be persisted in the MemStore

#### Scenario: Recall tool retrieves memory
- **WHEN** `memory_agent_recall` is called with a query
- **THEN** it SHALL return matching entries from the MemStore

#### Scenario: Forget tool removes memory
- **WHEN** `memory_agent_forget` is called with a key
- **THEN** the entry SHALL be removed from the MemStore
