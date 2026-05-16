## MODIFIED Requirements

### Requirement: CLI health check command
The system SHALL provide a `lango health` CLI command that checks the gateway health endpoint without external dependencies.

#### Scenario: Failed health check does not emit success payload
- **WHEN** `lango health` returns a non-200 status or times out
- **THEN** it SHALL return an error without emitting the `ok` success payload
