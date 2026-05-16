## ADDED Requirements

### Requirement: CLI harness hygiene guards stay executable
Repository-level CLI test-harness regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: CLI harness regressions are rejected
- **WHEN** a CLI test reintroduces process-global stdio replacement or legacy shared exec helpers
- **THEN** an executable repository test SHALL fail
