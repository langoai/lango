## MODIFIED Requirements

### Requirement: Knowledge relevance score mutations
The system SHALL provide `BoostRelevanceScore(ctx, key, delta, maxScore)` that atomically increases relevance_score for the latest version, capped at maxScore. It SHALL provide `DecayAllRelevanceScores(ctx, delta, minScore)` that subtracts delta from all latest-version entries with score > minScore + delta. It SHALL provide `ResetAllRelevanceScores(ctx)` that sets all latest-version scores to 1.0.

#### Scenario: Boost with cap
- **WHEN** BoostRelevanceScore is called and current score < maxScore
- **THEN** score SHALL increase by delta

#### Scenario: Boost at cap
- **WHEN** BoostRelevanceScore is called and current score >= maxScore
- **THEN** no update SHALL occur

#### Scenario: Decay with floor
- **WHEN** DecayAllRelevanceScores is called
- **THEN** only entries with score > minScore + delta SHALL be updated (prevents undershoot)

#### Scenario: Reset all
- **WHEN** ResetAllRelevanceScores is called
- **THEN** all latest-version entries SHALL have relevance_score set to 1.0
