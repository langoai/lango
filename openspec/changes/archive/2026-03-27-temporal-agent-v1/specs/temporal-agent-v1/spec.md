## ADDED Requirements

### Requirement: TemporalSearchAgent type
The `retrieval` package SHALL provide a `TemporalSearchAgent` struct that implements the `RetrievalAgent` interface. Its `Name()` SHALL return `"temporal-search"`. Its `Layers()` SHALL return `[LayerUserKnowledge]`.

#### Scenario: Agent identity
- **WHEN** TemporalSearchAgent is created
- **THEN** Name() SHALL return "temporal-search"
- **AND** Layers() SHALL return exactly [LayerUserKnowledge]

### Requirement: TemporalSearchSource interface
The `retrieval` package SHALL define a `TemporalSearchSource` interface with method: `SearchRecentKnowledge(ctx context.Context, query string, limit int) ([]knowledge.KnowledgeEntry, error)`. This interface SHALL be satisfied by `*knowledge.Store`.

### Requirement: Recency score normalization
TemporalSearchAgent SHALL compute recency scores as `max(0, 1.0 - hoursSinceUpdate / 168)` where 168 is the decay window in hours (1 week). Score range SHALL be 0.0 to 1.0.

#### Scenario: Recently updated entry
- **WHEN** an entry was updated 1 hour ago
- **THEN** its recency score SHALL be approximately 0.994

#### Scenario: Expired entry
- **WHEN** an entry was updated more than 168 hours ago
- **THEN** its recency score SHALL be 0.0

### Requirement: Content enrichment
TemporalSearchAgent SHALL prepend version and recency metadata to finding content in the format `[vN | updated <age>] <original content>`. Age format SHALL adapt: `just now` (<1min), `Xm ago` (<1h), `Xh ago` (<24h), `Xd ago` (>=24h).

#### Scenario: Content format
- **WHEN** an entry at version 5 was updated 3 hours ago
- **THEN** finding content SHALL start with `[v5 | updated 3h ago]`

### Requirement: SearchSource field
All findings from TemporalSearchAgent SHALL have `SearchSource = "temporal"`.

### Requirement: Injectable clock
TemporalSearchAgent SHALL support an injectable `now` function for deterministic testing. Default SHALL be `time.Now`.

### Requirement: KnowledgeEntry.UpdatedAt
`KnowledgeEntry` domain type SHALL include an `UpdatedAt time.Time` field. This field SHALL be populated from the Ent entity's `updated_at` field in all store query methods that return `KnowledgeEntry`.

#### Scenario: UpdatedAt populated
- **WHEN** GetKnowledge, GetKnowledgeHistory, SearchKnowledge, SearchKnowledgeScored, or SearchRecentKnowledge returns entries
- **THEN** each entry's UpdatedAt SHALL be populated from the database

### Requirement: SearchRecentKnowledge store method
`knowledge.Store` SHALL provide `SearchRecentKnowledge(ctx, query, limit)` that returns latest-version entries ordered by `updated_at DESC`. When query is non-empty, results SHALL be filtered by keyword matching (same predicates as existing LIKE search). The method SHALL NOT use FTS5 ordering.

#### Scenario: Recency ordering
- **WHEN** SearchRecentKnowledge is called
- **THEN** results SHALL be ordered by updated_at descending (most recent first)

#### Scenario: Keyword filtering
- **WHEN** SearchRecentKnowledge is called with a non-empty query
- **THEN** only entries matching at least one keyword SHALL be returned

### Requirement: v1 layer coverage boundary
The v1 TemporalSearchAgent SHALL cover only LayerUserKnowledge. AgentLearnings are excluded because learning entries lack version chains (no version/is_latest fields).
