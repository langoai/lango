## ADDED Requirements

### Requirement: Collaboration runtime seam regressions remain deterministic
Tests that temporarily replace the package-global collaboration runtime `eventTime` seam SHALL avoid parallel execution so sibling tests cannot observe the override unexpectedly.

#### Scenario: Event-time seam override does not leak
- **WHEN** a collaboration runtime regression overrides `eventTime`
- **THEN** repository-wide test runs SHALL not depend on whether sibling tests were scheduled at the same time
