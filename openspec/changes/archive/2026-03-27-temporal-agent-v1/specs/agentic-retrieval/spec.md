## MODIFIED Requirements

### Requirement: TemporalSearchAgent registration in coordinator
`initRetrievalCoordinator` SHALL always register `TemporalSearchAgent` alongside `FactSearchAgent`. Unlike `ContextSearchAgent` (which requires RAGService), `TemporalSearchAgent` has no optional dependencies — it uses `kStore` which is always available.

#### Scenario: Coordinator without RAG
- **WHEN** embedding components are nil
- **THEN** coordinator SHALL have 2 agents registered (fact-search + temporal-search)

#### Scenario: Coordinator with RAG
- **WHEN** embedding components with ragService are provided
- **THEN** coordinator SHALL have 3 agents registered (fact-search + temporal-search + context-search)
