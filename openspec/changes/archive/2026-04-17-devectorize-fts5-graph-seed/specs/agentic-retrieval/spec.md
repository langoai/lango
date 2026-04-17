## MODIFIED Requirements

### Requirement: RetrievalAgent interface
The `retrieval` package SHALL define a `RetrievalAgent` interface with methods: `Name() string`, `Layers() []knowledge.ContextLayer`, `Search(ctx context.Context, query string, limit int) ([]Finding, error)`. The `limit` parameter is item count, NOT token budget.

#### Scenario: Retrieval agent receives item-count limit
- **WHEN** the coordinator invokes a retrieval agent
- **THEN** the agent SHALL interpret `limit` as a maximum item count rather than a token budget

### Requirement: Layer coverage boundary
The coordinator SHALL cover 3 factual layers: UserKnowledge, AgentLearnings, ExternalKnowledge. ToolRegistry, RuntimeContext, SkillPatterns, PendingInquiries are handled by the ContextRetriever. The coordinator runs as primary in Phase 1 of GenerateContent (not shadow).

#### Scenario: Coordinator limits itself to factual layers
- **WHEN** the retrieval coordinator is initialized
- **THEN** it SHALL only return findings for UserKnowledge, AgentLearnings, and ExternalKnowledge
- **AND** it SHALL not retrieve ToolRegistry, RuntimeContext, SkillPatterns, or PendingInquiries layers

### Requirement: Configuration
`RetrievalConfig` SHALL have `Enabled` (bool, default false) and `Feedback` (bool, default false) fields, plus nested `AutoAdjust AutoAdjustConfig`. The coordinator SHALL only be created when `Enabled=true`.

#### Scenario: Coordinator disabled by config
- **WHEN** `retrieval.enabled` is `false`
- **THEN** the retrieval coordinator SHALL not be created

### Requirement: RAG enabled flag enforcement
The agentic retrieval coordinator SHALL NOT register a vector-backed context search agent. The coordinator SHALL operate without any embedding-specific dependency.

#### Scenario: Coordinator created with retrieval enabled
- **WHEN** the retrieval coordinator is initialized
- **THEN** it SHALL not require an embedding service to be present

### Requirement: TemporalSearchAgent registration in coordinator
`initRetrievalCoordinator` SHALL always register `TemporalSearchAgent` alongside `FactSearchAgent`. `TemporalSearchAgent` has no optional dependencies — it uses `kStore` which is always available.

#### Scenario: Temporal search always registered
- **WHEN** the retrieval coordinator is initialized with retrieval enabled
- **THEN** the coordinator SHALL include `TemporalSearchAgent`

### Requirement: ContextSearchAgent registration in coordinator
`initRetrievalCoordinator` SHALL register only `FactSearchAgent` and `TemporalSearchAgent`. The coordinator SHALL run those agents in parallel and SHALL NOT register the removed vector-backed `ContextSearchAgent`.

#### Scenario: Retrieval coordinator agent count
- **WHEN** the retrieval coordinator is initialized with retrieval enabled
- **THEN** the coordinator SHALL have 2 agents registered (fact-search + temporal-search)
