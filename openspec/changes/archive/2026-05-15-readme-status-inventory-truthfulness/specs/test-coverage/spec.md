## ADDED Requirements

### Requirement: README status inventory guard stays executable
Repository-level regressions that reintroduce vague status-family shorthand into the README internal CLI inventory SHALL be enforced by an executable test.

#### Scenario: Vague status shorthand is rejected
- **WHEN** the README internal tree still documents the shipped dead-letter command surface
- **THEN** an executable repository test SHALL fail if it falls back to vague wording instead of `lango status/dead-letter-summary/dead-letters/dead-letter/dead-letter retry`
