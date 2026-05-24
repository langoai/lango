## ADDED Requirements

### Requirement: README agent inventory slash-form guard stays executable
Repository-level regressions that reintroduce stale hyphen-compressed agent diagnostics wording into the README internal CLI inventory SHALL be enforced by an executable test.

#### Scenario: Stale hyphen shorthand is rejected
- **WHEN** the README internal tree still documents the shipped agent diagnostics surface
- **THEN** an executable repository test SHALL fail if it falls back to `trace list-show-metrics/graph` instead of `trace list/show/metrics/graph`
