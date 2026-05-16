## ADDED Requirements

### Requirement: README memory placeholder guard stays executable
Repository-level regressions that drop the `agent <name>` placeholder from the README internal memory inventory SHALL be enforced by an executable test.

#### Scenario: Placeholder loss is rejected
- **WHEN** the README internal tree still documents the shipped `lango memory agent <name>` surface
- **THEN** an executable repository test SHALL fail if it falls back to wording that omits the `<name>` placeholder
