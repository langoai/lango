## ADDED Requirements

### Requirement: Knowledge-artifact meta tools keep actionable missing-parameter errors

Knowledge save/exportability/release-approval tools SHALL reject missing required inputs at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: Knowledge-artifact tools reject missing required inputs
- **WHEN** `save_knowledge`, `evaluate_exportability`, or `approve_artifact_release` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
