## ADDED Requirements

### Requirement: Quickstart describes the workbench/cockpit split correctly
The public quickstart guide SHALL describe bare `lango` and `lango cockpit` using the current workbench/cockpit entrypoint split.

#### Scenario: Quickstart TUI tip uses current entrypoints
- **WHEN** a user reads the interactive TUI tip in `docs/getting-started/quickstart.md`
- **THEN** it SHALL describe bare `lango` as the standalone mission workbench
- **AND** it SHALL describe `lango cockpit` as the explicit multi-panel operator dashboard
