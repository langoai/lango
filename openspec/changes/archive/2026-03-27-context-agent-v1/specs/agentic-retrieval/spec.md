## ADDED Requirements

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
