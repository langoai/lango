## ADDED Requirements

### Requirement: CLI index core/status section guard stays executable
Repository-level regressions that remove dedicated core or status sections from the public CLI index SHALL be enforced by an executable test.

#### Scenario: Implemented core and status sections remain listed
- **WHEN** the repository still ships the implemented core command family and the `lango status` dead-letter command family
- **THEN** an executable repository test SHALL fail if `docs/cli/index.md` no longer includes dedicated `Core Commands` and `Status Dashboard` sections for them
