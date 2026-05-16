## ADDED Requirements

### Requirement: Skill read/import wrapper guards stay actionable

The skill read/import tools SHALL preserve actionable missing-parameter errors for all required wrapper inputs.

#### Scenario: View and import skill tools reject missing required inputs
- **WHEN** `view_skill` or `import_skill` is invoked without one of its required inputs
- **THEN** the tool SHALL return an actionable missing-parameter error before service execution begins
