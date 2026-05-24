## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Memory CLI tests avoid global stdout interception
- **WHEN** the memory CLI regression suite runs
- **THEN** it SHALL capture command output through package-local command writer helpers instead of a process-global stdout interception utility
