## ADDED Requirements

### Requirement: Public docs describe the two-step Enter quick-start flow

Public docs that describe ready-profile `Enter` quick-start behavior SHALL mention both the seed step and the submit step.

#### Scenario: Docs mention Enter seed then submit flow
- **WHEN** README or CLI/TUI docs describe the ready-profile `Enter` quick-start path
- **THEN** they SHALL mention that the first `Enter` seeds the default starter prompt
- **AND** they SHALL mention that the next `Enter` submits it
