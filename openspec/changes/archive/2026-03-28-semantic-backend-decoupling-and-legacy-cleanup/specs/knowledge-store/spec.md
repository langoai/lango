## ADDED Requirements

### Requirement: VectorStore build-tag isolation
The `sqlite-vec` backend SHALL be compiled only when the `vec` build tag is present. The default build (no `vec` tag) SHALL produce a binary without sqlite-vec linked. The `NewVectorStore` factory function SHALL be the canonical entry point for VectorStore creation in wiring code.

#### Scenario: Build with vec tag
- **WHEN** the binary is built with `-tags vec`
- **THEN** `NewVectorStore(db, dimensions)` SHALL return a `*SQLiteVecStore` implementing `VectorStore`

#### Scenario: Build without vec tag
- **WHEN** the binary is built without the `vec` tag
- **THEN** `NewVectorStore(db, dimensions)` SHALL return `nil, ErrVecNotCompiled`

#### Scenario: Wiring graceful degradation
- **WHEN** `NewVectorStore` returns an error during embedding initialization
- **THEN** the system SHALL log a warning and return nil `embeddingComponents` with a FeatureStatus suggesting "rebuild with -tags vec"
- **AND** RAG, EmbeddingBuffer, and ContextSearchAgent SHALL NOT be registered

### Requirement: Coordinator primary retrieval
The `RetrievalCoordinator` SHALL run as the primary retrieval path for factual layers (UserKnowledge, AgentLearnings, ExternalKnowledge) in Phase 1 of `GenerateContent`. The old `ContextRetriever` SHALL handle non-factual layers only (RuntimeContext, ToolRegistry, SkillPatterns, PendingInquiries).

#### Scenario: Phase 1 parallel retrieval
- **WHEN** `GenerateContent` is called with a non-empty user query and both retriever and coordinator are configured
- **THEN** both SHALL run as parallel goroutines in Phase 1
- **AND** retriever SHALL request only non-factual layers
- **AND** coordinator SHALL retrieve factual layers

#### Scenario: Result merge
- **WHEN** both retriever and coordinator return results
- **THEN** `mergeRetrievalResults` SHALL combine Items maps from both sources
- **AND** TotalItems SHALL equal the sum of both results' items

#### Scenario: ContextInjectedEvent coverage
- **WHEN** `ContextInjectedEvent` is published after context assembly
- **THEN** `Items` SHALL contain items from the merged result (both factual and non-factual layers)

## REMOVED Requirements

### Requirement: Shadow comparison mode (REMOVED)
The `RetrievalCoordinator` no longer supports shadow mode. The `Shadow` field, `SetShadow()`, and `Shadow()` methods are removed. The `RetrievalConfig.Shadow` configuration field is removed. The `CompareShadowResults` function and `shadow.go` file are deleted.

### Requirement: assembleMemorySection convenience wrapper (REMOVED)
The `assembleMemorySection` method is removed (zero production callers). Memory assembly uses the split `retrieveMemoryData` + `formatMemorySection` path exclusively.

### Requirement: assembleRunSummarySection convenience wrapper (REMOVED)
The `assembleRunSummarySection` method is removed. The run summary cache now stores `[]RunSummaryContext` instead of formatted strings. `retrieveRunSummaryData` returns cached summaries directly.
