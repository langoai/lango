## ADDED Requirements

### Requirement: Provenance inventory subcommand-slice guard stays executable
Repository-level regressions that reintroduce broad family-only provenance shorthand into the inventory docs SHALL be enforced by an executable test.

#### Scenario: Broad provenance shorthand is rejected
- **WHEN** the inventory docs still describe the shipped provenance surface
- **THEN** an executable repository test SHALL fail if they fall back to broad family-only wording instead of the current subcommand slices
