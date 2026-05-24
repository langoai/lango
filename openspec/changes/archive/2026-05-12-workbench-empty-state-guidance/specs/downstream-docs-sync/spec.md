## ADDED Requirements

### Requirement: Workbench docs mention incomplete-profile setup guidance
Public workbench documentation SHALL describe that the empty workbench state now points incomplete profiles at setup and verification commands.

#### Scenario: README and CLI/TUI docs mention setup recovery path
- **WHEN** a user reads the README workbench section or the CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that incomplete profiles are guided toward `lango onboard`, `lango settings`, and `lango doctor`
