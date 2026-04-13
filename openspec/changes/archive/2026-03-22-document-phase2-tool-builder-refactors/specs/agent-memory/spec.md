## RENAMED Requirements

- FROM: `### Requirement: MemoryEntry type`
- TO: `### Requirement: Entry type`

- FROM: `### Requirement: In-memory MemStore`
- TO: `### Requirement: In-memory InMemoryStore`

## MODIFIED Requirements

### Requirement: Entry type
The `agentmemory` package SHALL define an `Entry` struct with fields: ID, AgentName, Scope, Kind, Key, Content, Confidence, UseCount, Tags, CreatedAt, UpdatedAt.

#### Scenario: Entry fields
- **WHEN** an `Entry` is created
- **THEN** it SHALL expose all required fields for agent-scoped memory storage

### Requirement: In-memory InMemoryStore
The package SHALL provide an `InMemoryStore` implementation using sync.RWMutex-protected maps. It SHALL support Save (upsert), Get, Search, SearchWithContext, SearchWithContextOptions, Delete, IncrementUseCount, and Prune operations.

#### Scenario: Save upserts by key
- **WHEN** Save is called with an existing key
- **THEN** the entry SHALL be updated without duplication

#### Scenario: Search by agent and tags
- **WHEN** Search is called with agent name and tags
- **THEN** it SHALL return all matching entries for that agent

### Requirement: Agent memory tools
The `agentmemory` package SHALL export `BuildTools(store Store)` that returns `memory_agent_save`, `memory_agent_recall`, and `memory_agent_forget`. App-layer registration SHALL use this builder instead of `internal/app/tools_agentmemory.go`.

#### Scenario: BuildTools returns the three agent memory tools
- **WHEN** `agentmemory.BuildTools(store)` is called
- **THEN** it SHALL return tools named `memory_agent_save`, `memory_agent_recall`, and `memory_agent_forget`

#### Scenario: Recall tool preserves context fallback with kind filter
- **WHEN** `memory_agent_recall` is called with a query and a valid `kind`
- **THEN** it SHALL resolve results through `SearchWithContextOptions`
- **AND** matching global-scope entries SHALL remain eligible after kind filtering

## ADDED Requirements

### Requirement: SearchWithContextOptions filters before truncation
The `InMemoryStore` SHALL expose `SearchWithContextOptions(agentName, opts)` so context-aware lookup can apply `Kind`, `Tags`, and `MinConfidence` during collection, before the final result limit is applied.

#### Scenario: Kind filter still returns matching global entries under small limits
- **WHEN** instance-scoped entries of non-matching kinds outnumber the requested limit
- **AND** a global entry matches both the query and requested kind
- **THEN** `SearchWithContextOptions` SHALL still return the matching global entry

#### Scenario: Context-aware filters support tags and confidence
- **WHEN** `SearchWithContextOptions` is called with `Tags` and `MinConfidence`
- **THEN** only entries matching the query and those additional filters SHALL be returned

### Requirement: MemoryKind values are validated at runtime
The package SHALL expose `MemoryKind.Valid()` and reject invalid kind values at both tool-handler and store boundaries.

#### Scenario: Save tool rejects invalid kind
- **WHEN** `memory_agent_save` receives a `kind` value outside `pattern`, `preference`, `fact`, or `skill`
- **THEN** it SHALL return an error
- **AND** it SHALL NOT persist the entry

#### Scenario: Recall tool rejects invalid kind
- **WHEN** `memory_agent_recall` receives an invalid `kind`
- **THEN** it SHALL return an error instead of searching

#### Scenario: InMemoryStore Save rejects invalid kind
- **WHEN** `InMemoryStore.Save()` receives an entry whose `Kind` is non-empty and invalid
- **THEN** it SHALL return an error
- **AND** it SHALL NOT write the entry
