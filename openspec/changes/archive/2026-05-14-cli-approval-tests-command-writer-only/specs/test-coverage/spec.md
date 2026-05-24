## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Approval CLI tests avoid global stdout interception
- **WHEN** the approval CLI regression suite runs
- **THEN** it SHALL capture command output through the command writer helper instead of a process-global stdout interception utility
