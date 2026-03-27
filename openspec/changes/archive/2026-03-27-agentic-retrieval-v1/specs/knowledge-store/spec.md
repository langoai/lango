## MODIFIED Requirements

### Requirement: Knowledge Scored Search
The system SHALL provide `SearchKnowledgeScored(ctx, query, category, limit)` returning `[]ScoredKnowledgeEntry` with normalized Score (higher=better) and SearchSource ("fts5"/"like"). FTS5 path SHALL negate BM25 rank for normalization. LIKE path SHALL use RelevanceScore. Existing `SearchKnowledge` SHALL remain unchanged.

#### Scenario: FTS5 scored search
- **WHEN** `SearchKnowledgeScored` is called and FTS5 index is available
- **THEN** results SHALL include `Score = -rank` (BM25 negated) and `SearchSource = "fts5"`

#### Scenario: LIKE scored search
- **WHEN** `SearchKnowledgeScored` is called and FTS5 is unavailable
- **THEN** results SHALL include `Score = RelevanceScore` and `SearchSource = "like"`

#### Scenario: Scored search returns latest only
- **WHEN** `SearchKnowledgeScored` is called
- **THEN** only `is_latest=true` entries SHALL be returned

### Requirement: Learning Scored Search
The system SHALL provide `SearchLearningsScored(ctx, errorPattern, category, limit)` returning `[]ScoredLearningEntry` with Score and SearchSource.

#### Scenario: Learning scored search
- **WHEN** `SearchLearningsScored` is called
- **THEN** results SHALL include `Score = Confidence` and `SearchSource = "like"`
