## ADDED Requirements

### Requirement: Ontology governance and action tools keep actionable wrapper parameter guards

Ontology governance and dynamic action tools SHALL reject missing required wrapper inputs with actionable parameter errors before downstream ontology operations begin.

#### Scenario: Ontology governance and action tools reject missing required inputs
- **WHEN** `ontology_promote_type`, `ontology_promote_predicate`, `ontology_type_usage`, or any `ontology_action_*` tool is invoked without one of its declared required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not proceed into downstream ontology service execution
