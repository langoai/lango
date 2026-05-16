## ADDED Requirements

### Requirement: Repository test-harness guards stay executable
Repository-level test-harness regressions that are cheap to detect mechanically SHALL be enforced by executable tests instead of relying only on manual review.

#### Scenario: Repository test harness regressions are rejected
- **WHEN** a repository test reintroduces global stdio reassignment or legacy shared exec helpers
- **THEN** an executable repository test SHALL fail
