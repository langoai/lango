## Purpose

Capability spec for graph-rag. See requirements below for scope and behavior contracts.
## Requirements
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

### Requirement: LLM-based entity extraction
The system SHALL use an LLM to extract entities and relationships from saved knowledge and memory content, producing triples for the graph store.

#### Scenario: Entity extraction on knowledge save
- **WHEN** a knowledge entry is saved with `graph.enabled: true`
- **THEN** an async goroutine SHALL extract entities via the Extractor and enqueue resulting triples to GraphBuffer

#### Scenario: No meaningful relationships
- **WHEN** the LLM returns "NONE" for a content piece
- **THEN** no triples SHALL be enqueued (only the basic Contains triple from the direct callback)

### Requirement: Context injection for Graph RAG
The ContextAwareModelAdapter SHALL inject Graph RAG results into the system prompt when both graph store and GraphRAG are enabled.

#### Scenario: Graph RAG section in system prompt
- **WHEN** a query triggers GraphRAG retrieval with graph enabled
- **THEN** the system prompt SHALL include both "Retrieved Context" and "Graph-Expanded Context" sections

### Requirement: VectorRetrieveOptions supports MaxDistance
VectorRetrieveOptions SHALL include a MaxDistance field that is passed through to the underlying vector retrieval.

#### Scenario: MaxDistance passed to vector retriever
- **WHEN** Graph RAG retrieval is invoked with MaxDistance set
- **THEN** the MaxDistance value SHALL be forwarded to the VectorRetriever's RetrieveOptions

### Requirement: GraphNode carries entity type information
GraphNode SHALL include a `NodeType string` field populated from the discovered triple's SubjectType metadata. AssembleSection SHALL format typed nodes as `**NodeType:ID**` when NodeType is non-empty.

#### Scenario: GraphNode populated with type from traversal
- **WHEN** graph expansion discovers a node from a triple with `SubjectType: "ErrorPattern"`
- **THEN** the resulting GraphNode has `NodeType: "ErrorPattern"`

#### Scenario: GraphNode from untyped triple
- **WHEN** graph expansion discovers a node from a triple with empty SubjectType
- **THEN** the resulting GraphNode has `NodeType: ""` and is formatted as `**ID**` (no type prefix)

