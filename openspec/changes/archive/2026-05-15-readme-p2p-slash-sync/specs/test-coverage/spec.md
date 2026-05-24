## ADDED Requirements

### Requirement: README P2P slash-form guard stays executable
Repository-level regressions that reintroduce hyphen-compressed P2P subcommand slices into the README internal inventory SHALL be enforced by an executable test.

#### Scenario: Hyphen-compressed P2P shorthand is rejected
- **WHEN** the README internal tree still documents the shipped P2P surface
- **THEN** an executable repository test SHALL fail if it falls back to hyphen-compressed subcommand slices
