## MODIFIED Requirements

### Requirement: Two-phase hybrid retrieval
The GraphRAGService SHALL perform 2-phase retrieval: Phase 1 content retrieval, Phase 2 graph expansion (BFS traversal from Phase 1 results). The phase-1 retriever SHALL be supplied via a generic `ContentRetriever` interface so the seed backend can be FTS5-backed knowledge search instead of vector search.

#### Scenario: Retrieved content results expanded via graph
- **WHEN** a query returns phase-1 content results with source IDs matching graph nodes
- **THEN** the service SHALL traverse the graph from each result node and append discovered nodes as `GraphNode` entries

#### Scenario: No graph store available
- **WHEN** the graph store is nil or phase-1 content results are empty
- **THEN** the service SHALL return the phase-1 content results only without graph expansion

#### Scenario: Expansion limit respected
- **WHEN** graph traversal yields more nodes than `maxExpand`
- **THEN** the service SHALL stop expanding and return at most `maxExpand` graph results

### Requirement: Context injection for Graph RAG
The ContextAwareModelAdapter SHALL inject Graph RAG results into the system prompt when both graph store and GraphRAG are enabled.

#### Scenario: Graph RAG section in system prompt
- **WHEN** a query triggers GraphRAG retrieval with graph enabled
- **THEN** the system prompt SHALL include both "Retrieved Context" and "Graph-Expanded Context" sections
