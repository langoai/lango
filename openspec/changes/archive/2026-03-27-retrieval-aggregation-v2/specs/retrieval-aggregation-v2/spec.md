## ADDED Requirements

### Requirement: Finding provenance metadata
The `Finding` struct SHALL include `Source` (string), `Tags` ([]string), `Version` (int), and `UpdatedAt` (time.Time) fields. These fields are optional — zero values mean no provenance available.

#### Scenario: FactSearchAgent populates provenance
- **WHEN** FactSearchAgent creates a knowledge Finding
- **THEN** Source, Tags, Version, UpdatedAt SHALL be populated from ScoredKnowledgeEntry.Entry

#### Scenario: ContextSearchAgent has no provenance
- **WHEN** ContextSearchAgent creates a Finding
- **THEN** Source, Tags, Version, UpdatedAt SHALL be zero values (RAGResult lacks provenance)

### Requirement: Source authority ranking
The `retrieval` package SHALL define a `sourceAuthority` map ranking knowledge sources: `"knowledge"` (4), `"session_learning"` (3), `"proactive_librarian"` (2), `"conversation_analysis"` (1), `"memory"` (1), `"learning"` (1). Unknown/empty source SHALL have implicit authority 0.

### Requirement: Evidence-based merge
`mergeFindings` SHALL replace `dedupFindings`. For same `(Layer, Key)`, the winner SHALL be selected by deterministic priority chain: authority → version (supersedes) → recency (UpdatedAt) → score. When all provenance fields are empty, merge SHALL fall through to Score tiebreaker (preserving backward-compatible behavior).

#### Scenario: Authority wins over score
- **WHEN** two findings have the same (Layer, Key) with Source="knowledge" (score 0.3) and Source="conversation_analysis" (score 5.0)
- **THEN** the finding with Source="knowledge" SHALL win (authority 4 > 1)

#### Scenario: Version supersedes at same authority
- **WHEN** two findings have same Source but different Versions
- **THEN** the finding with higher Version SHALL win

#### Scenario: Recency at same authority and version
- **WHEN** two findings have same Source and Version but different UpdatedAt
- **THEN** the finding with more recent UpdatedAt SHALL win

#### Scenario: Score tiebreaker for empty provenance
- **WHEN** two findings have all-empty provenance fields
- **THEN** the finding with higher Score SHALL win

#### Scenario: Learning findings fallback
- **WHEN** learning-layer findings have empty Source/Version (LearningEntry lacks these)
- **THEN** merge SHALL fall through to Score tiebreaker

### Requirement: compareFindingPriority function
The `retrieval` package SHALL expose a `compareFindingPriority(a, b Finding) int` function that returns >0 if a is preferred, <0 if b preferred, 0 if equal.

### Requirement: Merge resolution vs global ranking separation
Merge resolution (authority-first) determines which variant of the SAME key survives. Global ranking (`sortFindingsByScore`) uses Score to order ALL surviving findings. These are separate concerns.

### Requirement: save_knowledge default source
The `save_knowledge` tool handler SHALL default the `source` parameter to `"knowledge"` (not `""`), ensuring user-explicit saves rank highest in authority. No backfill of existing data with empty Source.
