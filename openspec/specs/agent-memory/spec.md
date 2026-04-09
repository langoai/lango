## Purpose

Capability spec for agent-memory. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Entry type
The `agentmemory` package SHALL define an `Entry` struct with fields: ID, AgentName, Scope, Kind, Key, Content, Confidence, UseCount, Tags, CreatedAt, UpdatedAt.

#### Scenario: Entry fields
- **WHEN** an `Entry` is created
- **THEN** it SHALL have all required fields for agent-scoped memory storage

### Requirement: MemoryScope resolution
Memory lookups SHALL follow scope resolution order: instance (specific agent instance) > type (agent type) > global (all agents). Higher-priority scopes SHALL override lower ones.

#### Scenario: Instance scope takes priority
- **WHEN** a memory key exists at both instance and type scope
- **THEN** the instance-scope entry SHALL be returned

#### Scenario: Fallback to global scope
- **WHEN** a memory key exists only at global scope
- **THEN** the global-scope entry SHALL be returned

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

### Requirement: In-memory InMemoryStore
The package SHALL provide an `InMemoryStore` implementation using sync.RWMutex-protected maps. It SHALL support Save (upsert), Get, Search, SearchWithContext, SearchWithContextOptions, Delete, IncrementUseCount, and Prune operations.

#### Scenario: Save upserts by key
- **WHEN** Save is called with an existing key
- **THEN** the entry SHALL be updated (not duplicated)

#### Scenario: Search by agent and tags
- **WHEN** Search is called with agent name and tags
- **THEN** it SHALL return all matching entries sorted by use count descending

#### Scenario: SearchWithContextOptions applies kind before limit
- **WHEN** `SearchWithContextOptions` is called with query, kind, and limit
- **THEN** it SHALL apply the kind filter while collecting context-aware results
- **AND** it SHALL apply the final limit after those filters

### Requirement: Agent memory tools
The `agentmemory` package SHALL export `BuildTools(store Store)` that returns three tools: `memory_agent_save`, `memory_agent_recall`, and `memory_agent_forget`. App-layer registration SHALL use this builder instead of `internal/app/tools_agentmemory.go`.

#### Scenario: BuildTools returns the three agent memory tools
- **WHEN** `agentmemory.BuildTools(store)` is called
- **THEN** it SHALL return tools named `memory_agent_save`, `memory_agent_recall`, and `memory_agent_forget`

#### Scenario: Save tool stores memory
- **WHEN** `memory_agent_save` is called with key and content
- **THEN** the entry SHALL be persisted in the MemStore

#### Scenario: Recall tool retrieves memory
- **WHEN** `memory_agent_recall` is called with a query
- **THEN** it SHALL return matching entries from the MemStore

#### Scenario: Recall tool preserves scope fallback with kind filter
- **WHEN** `memory_agent_recall` is called with a query and a valid `kind`
- **THEN** it SHALL resolve results through `SearchWithContextOptions`
- **AND** matching global-scope entries SHALL remain eligible after kind filtering

#### Scenario: Forget tool removes memory
- **WHEN** `memory_agent_forget` is called with a key
- **THEN** the entry SHALL be removed from the MemStore

### Requirement: MemoryKind values are validated at runtime
The package SHALL expose `MemoryKind.Valid()` and reject invalid kind values at both tool-handler and store boundaries.

#### Scenario: Save tool rejects invalid kind
- **WHEN** `memory_agent_save` receives a `kind` outside `pattern`, `preference`, `fact`, or `skill`
- **THEN** it SHALL return an error
- **AND** it SHALL NOT persist the entry

#### Scenario: Recall tool rejects invalid kind
- **WHEN** `memory_agent_recall` receives an invalid `kind`
- **THEN** it SHALL return an error instead of searching

#### Scenario: InMemoryStore Save rejects invalid kind
- **WHEN** `InMemoryStore.Save()` receives an entry whose `Kind` is non-empty and invalid
- **THEN** it SHALL return an error
- **AND** it SHALL NOT write the entry
