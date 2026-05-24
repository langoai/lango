## ADDED Requirements

### Requirement: Knowledge-artifact wrapper guards stay actionable

The foundational knowledge-artifact tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: Knowledge-artifact tools reject missing required inputs
- **WHEN** `save_knowledge`, `evaluate_exportability`, or `approve_artifact_release` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
