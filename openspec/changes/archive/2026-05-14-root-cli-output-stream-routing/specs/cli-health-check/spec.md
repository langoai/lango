## MODIFIED Requirements

### Requirement: CLI health check command
The system SHALL provide a `lango health` CLI command that checks the gateway health endpoint without external dependencies.

#### Scenario: Successful health check writes to command output
- **WHEN** `lango health` succeeds
- **THEN** the `ok` response SHALL be written through the Cobra command output stream
