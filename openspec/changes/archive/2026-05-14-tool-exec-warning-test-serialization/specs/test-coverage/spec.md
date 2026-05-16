## ADDED Requirements

### Requirement: Exec warning seam regressions remain deterministic
Tests that temporarily replace the package-global `execWarningWriter` SHALL avoid parallel execution so suite results do not depend on test scheduling.

#### Scenario: Warning writer seam test does not race
- **WHEN** the exec warning regression temporarily replaces `execWarningWriter`
- **THEN** sibling tests SHALL not be able to observe that replacement concurrently
