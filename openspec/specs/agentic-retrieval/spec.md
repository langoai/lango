## ADDED Requirements

### Requirement: Finding type
The `retrieval` package SHALL provide a `Finding` struct with fields: Key (string), Content (string), Score (float64, higher=better), Category (string), SearchSource (string, "fts5"/"like"), Agent (string), Layer (knowledge.ContextLayer).

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
`RetrievalCoordinator` SHALL run all registered agents in parallel, merge findings, deduplicate by `(Layer, Key)` (keeping highest Score), and truncate to token budget.

#### Scenario: Parallel agent execution
- **WHEN** Retrieve is called with 2+ registered agents
- **THEN** all agents SHALL be invoked concurrently

#### Scenario: Dedup by (Layer, Key)
- **WHEN** two agents return findings with the same Layer and Key
- **THEN** only the finding with the highest Score SHALL be kept

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

### Requirement: Shadow mode comparison
When shadow=true, the coordinator SHALL run as a fire-and-forget goroutine after the existing retrieval path completes. `CompareShadowResults()` SHALL log overlap %, old-only count, and new-only count between old RetrievalResult and new findings.

#### Scenario: Shadow does not block LLM
- **WHEN** shadow mode is active
- **THEN** the coordinator goroutine SHALL NOT block the LLM call

### Requirement: v1 layer coverage boundary
The v1 coordinator SHALL cover only 3 factual layers: UserKnowledge, AgentLearnings, ExternalKnowledge. ToolRegistry, RuntimeContext, SkillPatterns, PendingInquiries remain handled by the existing ContextRetriever.

### Requirement: Configuration
`RetrievalConfig` SHALL have `Enabled` (bool, default false), `Shadow` (bool, default true), and `Feedback` (bool, default false) fields. The coordinator SHALL only be created when `Enabled=true`.

### Requirement: RetrievalConfig Feedback field
The `RetrievalConfig` struct SHALL include a `Feedback bool` field that enables context injection observability. This field SHALL operate independently of `Enabled` and `Shadow` — feedback observability SHALL work regardless of whether the agentic retrieval coordinator is enabled.

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

### Requirement: Shadow comparison factual-vs-new-context split
`CompareShadowResults` SHALL log both overall metrics AND factual-layer-only metrics. Factual layers are UserKnowledge, AgentLearnings, ExternalKnowledge.

#### Scenario: Factual split logged
- **WHEN** CompareShadowResults is called with findings from both FactSearchAgent and ContextSearchAgent
- **THEN** logs SHALL include `factual_overlap`, `factual_old_only`, `factual_new_only` alongside overall metrics

### Requirement: ContextSearchAgent registration in coordinator
`initRetrievalCoordinator` SHALL accept embedding components and register `ContextSearchAgent` when RAGService is available. The coordinator SHALL run both FactSearchAgent and ContextSearchAgent in parallel.

#### Scenario: RAG available
- **WHEN** embedding components with ragService are provided
- **THEN** coordinator SHALL have 2 agents registered (fact-search + context-search)

#### Scenario: RAG not available
- **WHEN** embedding components are nil
- **THEN** coordinator SHALL have 1 agent registered (fact-search only)
