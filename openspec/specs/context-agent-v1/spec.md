## ADDED Requirements

### Requirement: ContextSearchAgent
The system SHALL provide a `ContextSearchAgent` implementing `RetrievalAgent` that wraps `ContextSearchSource` for vector/semantic search. It SHALL cover `LayerUserKnowledge` and `LayerAgentLearnings` only in v1.

#### Scenario: Name
- **WHEN** `Name()` is called
- **THEN** it SHALL return `"context-search"`

#### Scenario: Layers
- **WHEN** `Layers()` is called
- **THEN** it SHALL return `[LayerUserKnowledge, LayerAgentLearnings]`

#### Scenario: Search with knowledge results
- **WHEN** RAGService returns results from "knowledge" collection
- **THEN** findings SHALL have `Layer=LayerUserKnowledge`, `SearchSource="vector"`, `Agent="context-search"`

#### Scenario: Search with learning results
- **WHEN** RAGService returns results from "learning" collection
- **THEN** findings SHALL have `Layer=LayerAgentLearnings`

#### Scenario: Observation/reflection filtered
- **WHEN** RAGService returns results from "observation" or "reflection" collection
- **THEN** those results SHALL be skipped (not converted to findings)

#### Scenario: Empty results
- **WHEN** RAGService returns no results
- **THEN** Search SHALL return an empty findings slice without error

### Requirement: ContextSearchSource interface
The system SHALL define `ContextSearchSource` interface with `Retrieve(ctx, query, opts) ([]embedding.RAGResult, error)`. `*embedding.RAGService` SHALL satisfy this interface.

#### Scenario: RAGService satisfies interface
- **WHEN** `*embedding.RAGService` is used as `ContextSearchSource`
- **THEN** the code SHALL compile without type assertion errors

### Requirement: Vector distance to score normalization
The system SHALL convert vector distance (lower=better) to score (higher=better) using `max(0, 1.0 - float64(distance))`.

#### Scenario: Perfect match
- **WHEN** distance is 0.0
- **THEN** score SHALL be 1.0

#### Scenario: Distant result
- **WHEN** distance is 1.5
- **THEN** score SHALL be 0.0 (floored)

### Requirement: Collection to layer mapping
The system SHALL map RAG collection names to context layers. Only "knowledge" and "learning" collections SHALL be mapped in v1. Other collections SHALL return `(0, false)`.

#### Scenario: Knowledge collection
- **WHEN** collection is "knowledge"
- **THEN** SHALL return `(LayerUserKnowledge, true)`

#### Scenario: Unknown collection
- **WHEN** collection is "observation" or "reflection"
- **THEN** SHALL return `(0, false)`
