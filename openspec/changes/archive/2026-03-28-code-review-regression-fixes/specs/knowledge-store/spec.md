## MODIFIED Requirements

### Requirement: Knowledge relevance score mutations (MODIFIED)
The system SHALL provide `BoostRelevanceScore(ctx, key, delta, maxScore)` that atomically increases relevance_score for the latest version, **clamped at maxScore** via two-step update (add where safe, cap remainder). It SHALL provide `DecayAllRelevanceScores(ctx, delta, minScore)` that subtracts delta from all latest-version entries, **floored at minScore** via two-step update (subtract where safe, floor remainder). It SHALL provide `ResetAllRelevanceScores(ctx)` that sets all latest-version scores to 1.0.

#### Scenario: Boost with cap (two-step clamping)
- **WHEN** BoostRelevanceScore is called and current score + delta would exceed maxScore
- **THEN** score SHALL be set to maxScore (not exceed it)

#### Scenario: Boost within range
- **WHEN** BoostRelevanceScore is called and current score + delta <= maxScore
- **THEN** score SHALL increase by delta

#### Scenario: Decay with floor (two-step clamping)
- **WHEN** DecayAllRelevanceScores is called and current score - delta would go below minScore
- **THEN** score SHALL be set to minScore (not go below it)

#### Scenario: Decay within range
- **WHEN** DecayAllRelevanceScores is called and current score - delta >= minScore
- **THEN** score SHALL decrease by delta

## ADDED Requirements

### Requirement: RAG enabled flag enforcement
The system SHALL NOT create `RAGService` or register `ContextSearchAgent` when `embedding.rag.enabled` is false. The embedding buffer and provider SHALL still be initialized for async knowledge embedding regardless of the RAG flag.

#### Scenario: RAG disabled with embedding configured
- **WHEN** `embedding.provider` is configured but `embedding.rag.enabled` is false
- **THEN** `ragService` SHALL be nil in embeddingComponents
- **AND** `ContextSearchAgent` SHALL NOT be registered in the coordinator

### Requirement: Settings TUI explicit key preservation
The system SHALL mark all context-related config keys as explicitly set when saving from the settings TUI. This prevents `ResolveContextAutoEnable` from overriding user intent on subsequent bootstrap.

#### Scenario: Settings save preserves disabled flags
- **WHEN** user sets `knowledge.enabled=false` in settings TUI and saves
- **THEN** the saved explicitKeys SHALL include `knowledge.enabled`
- **AND** subsequent bootstrap SHALL NOT auto-enable knowledge
