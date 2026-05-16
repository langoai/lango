## ADDED Requirements

### Requirement: Knowledge read wrapper guards stay actionable

The knowledge history/search tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Knowledge history and search reject missing required inputs
- **WHEN** `get_knowledge_history` or `search_knowledge` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
