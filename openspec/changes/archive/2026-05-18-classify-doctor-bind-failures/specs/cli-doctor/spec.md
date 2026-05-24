## MODIFIED Requirements

### Requirement: Server Port Check
The system SHALL verify that the configured server port is available using the same bracket-safe listen-address formatting as the gateway server, and SHALL distinguish occupied-port failures from bind-address configuration failures.

#### Scenario: Port available
- **WHEN** configured port (default 18789) is not in use
- **THEN** check passes with "Port 18789 available"

#### Scenario: Port in use
- **WHEN** configured port is already bound by another process
- **THEN** check fails with "Port 18789 in use" and process information if available

#### Scenario: Invalid bind host
- **WHEN** configured `server.host` cannot be used as a local bind address
- **THEN** check fails with "Server bind address unavailable"
- **AND** the diagnostic details preserve the original listen error

#### Scenario: IPv6 host port available
- **WHEN** configured `server.host` is an IPv6 literal and the configured port is not in use
- **THEN** the server port check SHALL use a bracket-safe listen address
- **AND** the check SHALL pass instead of failing due to malformed address formatting
