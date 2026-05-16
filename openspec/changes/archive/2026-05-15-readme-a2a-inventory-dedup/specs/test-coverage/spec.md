## ADDED Requirements

### Requirement: Duplicate A2A README inventory guard stays executable
Repository-level regressions that reintroduce duplicate `a2a/` rows into the README internal CLI inventory SHALL be enforced by an executable test.

#### Scenario: Duplicate A2A rows are rejected
- **WHEN** the README internal tree still documents the shipped `lango a2a card/check` surface
- **THEN** an executable repository test SHALL fail if that `a2a/` inventory row appears more than once
