## MODIFIED Requirements

### Requirement: GVisor stub behavior is documented and tested
The GVisor runtime stub SHALL clearly document its stub nature and have tests verifying stub behavior.

#### Scenario: Explicit gVisor runtime request returns actionable unavailable error
- **WHEN** the container executor is configured with runtime `gvisor`
- **AND** the current build still uses the gVisor stub
- **THEN** executor construction SHALL fail with an error that explains the requested gVisor runtime is unavailable

### Requirement: Container runtime fail-closed errors stay actionable
Sandbox runtime selection SHALL keep operator-facing unavailable errors specific enough to explain which runtime or policy path failed.

#### Scenario: Explicit Docker runtime request names the unavailable runtime
- **WHEN** the container executor is configured with runtime `docker`
- **AND** Docker is unavailable in the current environment
- **THEN** executor construction SHALL fail with an error that explains the requested Docker runtime is unavailable

#### Scenario: Require-container fail-closed path names the container requirement
- **WHEN** the container executor is configured with `RequireContainer=true`
- **AND** no container runtime is available
- **THEN** executor construction SHALL fail with an error that explains container runtime availability is required rather than silently falling back to native execution
