## ADDED Requirements

### Requirement: Workbench docs mention state-aware composer guidance
Public workbench documentation SHALL mention that the composer placeholder follows the same incomplete-vs-ready guidance split as the empty-state body.

#### Scenario: README and CLI/TUI docs mention composer guidance split
- **WHEN** a user reads the README or CLI/TUI docs for the workbench surface
- **THEN** those docs SHALL mention that incomplete profiles get setup-first composer guidance and ready profiles get starter-prompt composer guidance
