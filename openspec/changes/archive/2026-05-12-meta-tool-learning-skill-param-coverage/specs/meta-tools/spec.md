## ADDED Requirements

### Requirement: Learning and skill management tools keep actionable missing-parameter errors

Learning and skill-management tools SHALL reject missing required inputs at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Learning and skill tools reject missing required inputs
- **WHEN** `save_learning`, `search_learnings`, or `create_skill` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
