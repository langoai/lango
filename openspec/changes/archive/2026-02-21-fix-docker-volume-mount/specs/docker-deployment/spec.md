## MODIFIED Requirements

### Requirement: Data Persistence
The system SHALL persist data across container restarts.

#### Scenario: SQLite database persistence
- **WHEN** the container restarts
- **THEN** the SQLite database at /home/lango/.lango SHALL be preserved
- **AND** session history SHALL not be lost

#### Scenario: Volume mount
- **WHEN** docker-compose is used
- **THEN** a named volume (lango-data) SHALL be mounted at /home/lango/.lango

### Requirement: Docker Container Configuration
The system SHALL provide a Dockerfile optimized for production deployment.

#### Scenario: Non-root execution
- **WHEN** the container starts
- **THEN** the lango process SHALL run as non-root user
- **AND** WORKDIR SHALL be `/home/lango` (user home directory, writable)
- **AND** the Dockerfile SHALL NOT create a separate `/data` directory
