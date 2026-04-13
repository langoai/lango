## MODIFIED Requirements

### Requirement: RetrievalCoordinator merge strategy
`RetrievalCoordinator` SHALL use evidence-based merge (`mergeFindings`) instead of score-only dedup (`dedupFindings`). The merge priority chain SHALL be: authority → version (supersedes) → recency → score. The coordinator SHALL still sort all surviving findings by Score descending and truncate to token budget.

#### Scenario: Dedup by (Layer, Key) with authority
- **WHEN** two agents return findings with the same Layer and Key but different Source authority
- **THEN** the finding with higher sourceAuthority SHALL be kept, regardless of Score
