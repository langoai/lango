## ADDED Requirements
### Requirement: CLI pretty-JSON writer guards stay executable
Repository-level CLI pretty-JSON writer regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Duplicate CLI pretty-JSON writer setups are rejected
- **WHEN** CLI production code reintroduces direct pretty-JSON indentation setup outside the shared CLI JSON helper
- **THEN** an executable repository test SHALL fail
