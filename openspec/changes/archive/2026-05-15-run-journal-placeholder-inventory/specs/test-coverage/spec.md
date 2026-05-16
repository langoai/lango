## ADDED Requirements

### Requirement: Run journal placeholder guard stays executable
Repository-level regressions that drop the `<run-id>` placeholder from the run inventory docs SHALL be enforced by an executable test.

#### Scenario: Placeholder loss is rejected
- **WHEN** the inventory docs still describe the shipped `lango run journal <run-id>` command
- **THEN** an executable repository test SHALL fail if they fall back to bare `journal`
