## ADDED Requirements

### Requirement: Ontology governance/action wrapper guards stay actionable

Ontology governance and dynamic action tools SHALL preserve actionable missing-parameter errors for all declared wrapper inputs.

#### Scenario: Ontology governance and action tools reject missing required inputs
- **WHEN** `ontology_promote_type`, `ontology_promote_predicate`, `ontology_type_usage`, or any `ontology_action_*` tool is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before downstream ontology execution begins
