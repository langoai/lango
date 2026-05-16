## MODIFIED Requirements

### Requirement: Existing Test Enhancements
Existing test files SHALL be expanded with additional scenarios.

#### Scenario: Prompt confirm helper covers yes/no/error branches
- **WHEN** the prompt package regression suite runs
- **THEN** `ConfirmIO(...)` SHALL be covered for approval, denial, and read-error paths
