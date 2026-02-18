## Purpose

Provide a built-in CLI health check command that eliminates the need for external tools like curl in Docker health checks.

## Requirements

### Requirement: CLI health check command
The system SHALL provide a `lango health` CLI command that checks the gateway health endpoint without external dependencies.

#### Scenario: Successful health check
- **WHEN** `lango health` is executed and the gateway is running on the default port
- **THEN** the system SHALL send an HTTP GET request to `http://localhost:18789/health`
- **AND** the system SHALL print "ok" and exit with code 0 when the response status is 200

#### Scenario: Failed health check
- **WHEN** `lango health` is executed and the gateway is not running or returns non-200
- **THEN** the system SHALL exit with code 1
- **AND** the system SHALL print an error message describing the failure

#### Scenario: Custom port
- **WHEN** `lango health --port 8080` is executed
- **THEN** the system SHALL check `http://localhost:8080/health` instead of the default port

#### Scenario: Request timeout
- **WHEN** `lango health` is executed and the gateway does not respond within 5 seconds
- **THEN** the system SHALL exit with code 1
