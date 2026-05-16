## ADDED Requirements

### Requirement: README provenance slash-form guard stays executable
Repository-level regressions that reintroduce hyphen-compressed provenance subcommand slices into the README internal inventory SHALL be enforced by an executable test.

#### Scenario: Hyphen-compressed provenance shorthand is rejected
- **WHEN** the README internal tree still documents the shipped provenance surface
- **THEN** an executable repository test SHALL fail if it falls back to hyphen-compressed subcommand slices
