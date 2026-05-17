## ADDED Requirements

### Requirement: Application Gateway Startup Failure Propagation
The application lifecycle SHALL propagate immediate gateway listener setup failures during startup. The `lango serve` command SHALL only print the startup summary after application startup succeeds.

#### Scenario: Occupied gateway port fails application startup
- **GIVEN** the configured gateway listen address is already occupied
- **WHEN** application startup starts lifecycle components
- **THEN** startup SHALL return an error for the gateway component
- **AND** it SHALL NOT report the application as successfully started

#### Scenario: Serve summary suppressed after startup failure
- **GIVEN** the application builder returns an application whose startup fails
- **WHEN** `lango serve` is executed
- **THEN** the command SHALL return the startup error
- **AND** it SHALL NOT print the startup summary

#### Scenario: Direct gateway start keeps blocking behavior
- **WHEN** `gateway.Server.Start()` is called directly
- **THEN** it SHALL bind the configured address and block while serving
- **AND** it SHALL still return `nil` after graceful shutdown
- **AND** it SHALL still return non-shutdown serve errors to the caller
