## ADDED Requirements

### Requirement: Learning and skill wrapper guards stay actionable

The learning and skill-management tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Learning and skill tools reject missing required inputs
- **WHEN** `save_learning`, `search_learnings`, or `create_skill` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
