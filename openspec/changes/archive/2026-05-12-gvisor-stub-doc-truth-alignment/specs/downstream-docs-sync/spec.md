## MODIFIED Requirements

### Requirement: Container runtime docs stay honest about the gVisor stub
Public configuration docs SHALL describe the current gVisor runtime status accurately instead of implying that a real gVisor backend is already available.

#### Scenario: Runtime tables mention gVisor stub status
- **WHEN** a user reads the runtime configuration tables in README or configuration docs
- **THEN** those tables SHALL still list `gvisor` as an accepted runtime value
- **AND** SHALL clarify that the current implementation is a stub whose explicit selection returns a runtime-unavailable error
