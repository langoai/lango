## ADDED Requirements

### Requirement: README P2P subcommand-slice guard stays executable
Repository-level regressions that reintroduce broad family-only shorthand into the README internal P2P inventory SHALL be enforced by an executable test.

#### Scenario: Broad P2P family-only shorthand is rejected
- **WHEN** the README internal tree still documents the shipped P2P surface
- **THEN** an executable repository test SHALL fail if it falls back to broad family-only wording instead of the current subcommand slices
