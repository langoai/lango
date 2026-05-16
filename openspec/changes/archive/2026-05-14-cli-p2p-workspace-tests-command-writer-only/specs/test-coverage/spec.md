## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: P2P workspace CLI tests avoid global stdout interception
- **WHEN** the P2P workspace CLI regression suite runs
- **THEN** it SHALL capture command output through the package-local command writer helper instead of a process-global stdout interception utility
