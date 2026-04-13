## MODIFIED Requirements

### Requirement: Knowledge relevance score mutations (MODIFIED — boost cap scope)
The `BoostRelevanceScore` cap step SHALL normalize ALL rows with `score > maxScore - delta`, including pre-existing over-cap values where `score >= maxScore`. This ensures that legacy over-cap data is corrected on the next boost call rather than persisting indefinitely.

#### Scenario: Pre-existing over-cap normalization
- **WHEN** BoostRelevanceScore is called and a row has `score > maxScore` (from prior bug)
- **THEN** score SHALL be set to maxScore
