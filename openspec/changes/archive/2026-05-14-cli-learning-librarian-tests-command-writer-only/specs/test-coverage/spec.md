## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Learning and librarian CLI tests avoid global stdout interception
- **WHEN** the learning and librarian CLI regression suites run
- **THEN** they SHALL capture command output through their package-local command writer helpers instead of a process-global stdout interception utility
