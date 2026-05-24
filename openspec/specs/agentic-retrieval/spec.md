## Purpose

Capability spec for agentic-retrieval. See requirements below for scope and behavior contracts.
## Requirements
### Requirement: Finding type
The `retrieval` package SHALL provide a `Finding` struct with fields: Key (string), Content (string), Score (float64, higher=better), Category (string), SearchSource (string, "fts5"/"like"/"vector"/"temporal"), Agent (string), Layer (knowledge.ContextLayer). The `SearchSource` field documents the retrieval METHOD used. The `Source` field documents the AUTHORSHIP origin.

#### Scenario: Finding from FTS5 search
- **WHEN** FactSearchAgent returns a finding from FTS5 path
- **THEN** Finding.Score SHALL be the negated BM25 rank (higher=better)
- **AND** Finding.SearchSource SHALL be "fts5"

#### Scenario: Finding from LIKE fallback
- **WHEN** FactSearchAgent returns a finding from LIKE path
- **THEN** Finding.Score SHALL be the RelevanceScore value
- **AND** Finding.SearchSource SHALL be "like"

### Requirement: RetrievalAgent interface
The `retrieval` package SHALL define a `RetrievalAgent` interface with methods: `Name() string`, `Layers() []knowledge.ContextLayer`, `Search(ctx context.Context, query string, limit int) ([]Finding, error)`. The `limit` parameter is item count, NOT token budget.

#### Scenario: Retrieval agent receives item-count limit
- **WHEN** the coordinator invokes a retrieval agent
- **THEN** the agent SHALL interpret `limit` as a maximum item count rather than a token budget

### Requirement: FactSearchAgent
`FactSearchAgent` SHALL implement `RetrievalAgent`. It SHALL depend on `FactSearchSource` interface (not concrete `*knowledge.Store`). It SHALL cover 3 factual layers: UserKnowledge, AgentLearnings, ExternalKnowledge. It SHALL call `SearchKnowledgeScored`, `SearchLearningsScored`, `SearchExternalRefs` and convert results to `[]Finding`.

#### Scenario: FactSearchAgent search
- **WHEN** Search is called with a query
- **THEN** it SHALL search knowledge, learnings, and external refs
- **AND** merge all results into a single []Finding slice

#### Scenario: External refs have zero score
- **WHEN** external refs are retrieved
- **THEN** Finding.Score SHALL be 0 and SearchSource SHALL be "like"

### Requirement: RetrievalCoordinator
`RetrievalCoordinator` SHALL run all registered agents in parallel, merge findings using evidence-based priority (authority → version → recency → score), and truncate to token budget. The merge SHALL use a pre-allocated `[][]Finding` slice indexed by agent position, with each goroutine writing to its own index. No mutex SHALL be required for the merge. The merge priority chain replaces score-only dedup: for same `(Layer, Key)`, the winner is selected by `sourceAuthority[Source]` first, then version (supersedes), then recency (UpdatedAt), then Score as final tiebreaker. When all provenance fields are empty, merge falls through to Score (backward compatible).

#### Scenario: Parallel retrieval with lock-free merge
- **WHEN** Retrieve is called with 2+ registered agents
- **THEN** each agent SHALL search concurrently
- **AND** results SHALL be merged without mutex contention
- **AND** agent errors SHALL be logged but not propagated (non-fatal)

#### Scenario: Dedup by (Layer, Key) with authority
- **WHEN** two agents return findings with the same Layer and Key but different Source authority
- **THEN** the finding with higher sourceAuthority SHALL be kept, regardless of Score

#### Scenario: Different layers same key preserved
- **WHEN** two findings have the same Key but different Layer values
- **THEN** both findings SHALL be kept

#### Scenario: Token budget truncation
- **WHEN** tokenBudget > 0 and merged findings exceed budget
- **THEN** coordinator SHALL drop lowest-score findings until within budget

### Requirement: ToRetrievalResult conversion
`RetrievalCoordinator.ToRetrievalResult()` SHALL convert `[]Finding` to `*knowledge.RetrievalResult`, grouping by Layer and setting `ContextItem.Score = Finding.Score`.

#### Scenario: Score propagated
- **WHEN** ToRetrievalResult converts findings
- **THEN** each ContextItem.Score SHALL equal the corresponding Finding.Score

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

### Requirement: RetrievalConfig Feedback field
The `RetrievalConfig` struct SHALL include a `Feedback bool` field that enables context injection observability. This field SHALL operate independently of `Enabled` — feedback observability SHALL work regardless of whether the agentic retrieval coordinator is enabled.

#### Scenario: Feedback enabled without coordinator
- **WHEN** `retrieval.feedback` is `true` and `retrieval.enabled` is `false`
- **THEN** the `FeedbackProcessor` SHALL be subscribed to the event bus

#### Scenario: Feedback default
- **WHEN** no `retrieval.feedback` value is configured
- **THEN** feedback SHALL default to `false`

### Requirement: Event bus wiring on ContextAwareModelAdapter
The `ContextAwareModelAdapter` SHALL accept an event bus via `WithEventBus(*eventbus.Bus)`. The event bus SHALL be wired unconditionally when a `ContextAwareModelAdapter` exists, regardless of coordinator presence.

#### Scenario: Bus wired in knowledge branch
- **WHEN** knowledge system is enabled and ctxAdapter is created
- **THEN** `WithEventBus(eventBus)` SHALL be called on the adapter

#### Scenario: Bus wired in OM-only branch
- **WHEN** only observational memory is enabled (no knowledge) and ctxAdapter is created
- **THEN** `WithEventBus(eventBus)` SHALL be called on the adapter

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

