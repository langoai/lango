## ADDED Requirements

### Requirement: Logging default writer seam regressions stay executable

Repository-level regressions that route default logs back to stdout or bypass the default logging writer seam SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Default logging writer seam is covered

- **WHEN** the logging package tests run
- **THEN** they SHALL fail if default logging output stops using the injected writer seam
