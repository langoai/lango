## ADDED Requirements

### Requirement: Workbench docs mention readiness-aware header summary
Public workbench documentation SHALL mention that the header summary now reports setup-required state for incomplete profiles.

#### Scenario: README and CLI/TUI docs mention setup-required header
- **WHEN** a user reads the README or CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that incomplete profiles show `Model: Setup required` in the header summary
