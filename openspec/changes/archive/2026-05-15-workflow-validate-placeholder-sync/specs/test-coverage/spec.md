## ADDED Requirements

### Requirement: Workflow validate placeholder guard stays executable
Repository-level regressions that drop the `<file>` placeholder from the workflow validate inventory docs SHALL be enforced by an executable test.

#### Scenario: Placeholder loss is rejected
- **WHEN** the inventory docs still describe the shipped workflow validate command
- **THEN** an executable repository test SHALL fail if they fall back to bare `validate`
