## MODIFIED Requirements

### Requirement: RetrievalCoordinator parallel merge
`RetrievalCoordinator.Retrieve()` SHALL run all agents in parallel via errgroup and merge results. The merge SHALL use a pre-allocated `[][]Finding` slice indexed by agent position, with each goroutine writing to its own index. No mutex SHALL be required for the merge.

#### Scenario: Parallel retrieval with lock-free merge
- **WHEN** `Retrieve` is called with multiple agents
- **THEN** each agent SHALL search concurrently
- **AND** results SHALL be merged without mutex contention
- **AND** agent errors SHALL be logged but not propagated (non-fatal)

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
