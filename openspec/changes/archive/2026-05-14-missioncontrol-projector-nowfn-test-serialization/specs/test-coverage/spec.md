## ADDED Requirements

### Requirement: Mission-control projector time-seam regressions remain deterministic
Tests that temporarily replace the mission-control projector `nowFn` seam SHALL avoid parallel execution so sibling tests cannot observe the override unexpectedly.

#### Scenario: nowFn seam override does not leak
- **WHEN** a mission-control projector regression overrides `nowFn`
- **THEN** repository-wide test runs SHALL not depend on whether sibling tests were scheduled at the same time
