## ADDED Requirements

### Requirement: README metrics inventory slash-form guard stays executable
Repository-level regressions that reintroduce stale bracket shorthand into the README internal metrics inventory SHALL be enforced by an executable test.

#### Scenario: Stale metrics bracket shorthand is rejected
- **WHEN** the README internal tree still documents the shipped metrics surface
- **THEN** an executable repository test SHALL fail if it falls back to bracket shorthand instead of `lango metrics/sessions/tools/agents/policy/history`
