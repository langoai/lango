## ADDED Requirements

### Requirement: Skill read/import tools keep actionable missing-parameter errors

Skill read/import tools SHALL reject missing required inputs at the wrapper layer with actionable parameter errors before invoking service logic.

#### Scenario: View and import skill tools reject missing required inputs
- **WHEN** `view_skill` is invoked without `name`
- **OR** `import_skill` is invoked without `url`
- **THEN** the tool SHALL return an actionable missing-parameter error
- **AND** SHALL not invoke the underlying service
