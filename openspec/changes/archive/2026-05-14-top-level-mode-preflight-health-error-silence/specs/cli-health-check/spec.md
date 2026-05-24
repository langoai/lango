## ADDED Requirements
### Requirement: Health command failures stay free of duplicate command output
`lango health` failures SHALL return an error without emitting the `ok` success payload or a duplicate Cobra-managed error body through the command output stream.

#### Scenario: Failed health check leaves command output empty
- **WHEN** `lango health` returns a non-200 status or times out
- **THEN** the command output stream SHALL remain empty
- **AND** the returned error SHALL still describe the failure
