## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Shared CLI stdout interception helper is removed when unused
- **WHEN** all in-repo CLI test call sites have been migrated to package-local command writer helpers
- **THEN** the unused shared global stdout interception helper SHALL be removed from `internal/testutil`
