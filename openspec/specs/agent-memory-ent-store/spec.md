## ADDED Requirements

### Requirement: Ent-backed persistent storage for agent memory
The system SHALL provide an Ent-backed implementation of `agentmemory.Store` that persists agent memory entries to SQLite via the existing Ent framework.

#### Scenario: EntStore created with Ent client
- **WHEN** `NewEntStore(client)` is called with a valid `*ent.Client`
- **THEN** the returned store SHALL implement all 11 `Store` interface methods

### Requirement: Save performs upsert by agent_name + key
The `Save()` method SHALL create a new entry if no entry exists for the given agent_name + key combination, or update the existing entry's mutable fields (content, confidence, kind, scope, tags).

#### Scenario: Save new entry
- **WHEN** `Save()` is called with agent_name="researcher" and key="pattern-1" for the first time
- **THEN** a new database row SHALL be created with the provided fields

#### Scenario: Save existing entry (upsert)
- **WHEN** `Save()` is called with an agent_name + key that already exists
- **THEN** the existing row SHALL be updated with new content, confidence, kind, scope, and tags
- **AND** `updated_at` SHALL be refreshed
- **AND** `created_at` and `use_count` SHALL be preserved

### Requirement: Search returns sorted results with filters
The `Search()` method SHALL match entries by keyword across content and key fields, apply Kind/Tags/MinConfidence filters, sort by confidence DESC, use_count DESC, updated_at DESC, and apply a limit.

#### Scenario: Keyword search with sort order
- **WHEN** `Search("researcher", SearchOptions{Query: "error"})` is called
- **THEN** results SHALL be sorted by confidence DESC, use_count DESC, updated_at DESC

### Requirement: SearchWithContext uses 2-phase scope resolution
The `SearchWithContext()` and `SearchWithContextOptions()` methods SHALL use 2-phase search: Phase 1 collects all entries for the given agent (any scope), Phase 2 collects global-scoped entries from other agents. Results are merged with Phase 1 first, then Phase 2, then global limit applied.

#### Scenario: Phase 1 returns agent's own entries
- **WHEN** `SearchWithContext("researcher", "error", 10)` is called
- **THEN** Phase 1 SHALL return all of researcher's entries matching the query (regardless of scope)

#### Scenario: Phase 2 returns other agents' global entries
- **WHEN** Phase 1 is complete
- **THEN** Phase 2 SHALL return entries from other agents WHERE scope = "global" AND matching the query

#### Scenario: Merge order preserves agent priority
- **WHEN** Phase 1 and Phase 2 results are merged
- **THEN** Phase 1 results SHALL appear before Phase 2 results
- **AND** within each phase, results SHALL be sorted by confidence DESC, use_count DESC, updated_at DESC

### Requirement: Agent name propagation via ToolAdapter
The `orchestration.ToolAdapter` type SHALL accept agent name as a parameter so that tool handlers receive the owning agent's name in context.

#### Scenario: Sub-agent tool receives agent name
- **WHEN** a sub-agent named "researcher" invokes `memory_agent_save`
- **THEN** the tool handler SHALL receive "researcher" from `ctxkeys.AgentNameFromContext(ctx)`

#### Scenario: Root agent tool receives agent name
- **WHEN** the root agent invokes `memory_agent_save`
- **THEN** the tool handler SHALL receive "lango-agent" from `ctxkeys.AgentNameFromContext(ctx)`
