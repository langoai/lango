## MODIFIED Requirements

### Requirement: GVisor stub behavior is documented and tested
The GVisor runtime stub SHALL clearly document its stub nature and have tests verifying stub behavior.

#### Scenario: Explicit gVisor runtime request returns actionable unavailable error
- **WHEN** the container executor is configured with runtime `gvisor`
- **AND** the current build still uses the gVisor stub
- **THEN** executor construction SHALL fail with an error that explains the requested gVisor runtime is unavailable
