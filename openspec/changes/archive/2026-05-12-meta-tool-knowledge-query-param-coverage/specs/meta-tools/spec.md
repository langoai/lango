## ADDED Requirements

### Requirement: Knowledge read tools keep actionable missing-parameter errors

Knowledge history/search tools SHALL reject missing required inputs at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Knowledge history and search reject missing required inputs
- **WHEN** `get_knowledge_history` is invoked without `key`
- **OR** `search_knowledge` is invoked without `query`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
