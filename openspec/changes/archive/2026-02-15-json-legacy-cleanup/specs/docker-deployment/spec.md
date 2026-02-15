## ADDED Requirements

### Requirement: Headless configuration via import
The system SHALL document a headless configuration pattern for Docker/CI environments where interactive setup is unavailable.

#### Scenario: Docker import workflow
- **WHEN** a Docker container needs configuration without interactive TUI
- **THEN** the user SHALL prepare a JSON config file, COPY it into the container, and run `lango config import config.json --profile production`
- **AND** the source JSON file is automatically deleted after import for security

#### Scenario: Non-interactive passphrase
- **WHEN** running in a headless environment without a terminal
- **THEN** the user SHALL set the `LANGO_PASSPHRASE` environment variable for non-interactive passphrase entry

## MODIFIED Requirements

### Requirement: Docker Compose Orchestration
The system SHALL provide a docker-compose.yml for simplified deployment. The README documentation SHALL describe the import→delete configuration pattern instead of read-only JSON mounting.

#### Scenario: Service definition
- **WHEN** running `docker-compose up`
- **THEN** the lango service SHALL start on port 18789
- **AND** volumes SHALL persist data to lango-data volume

#### Scenario: Configuration via import
- **WHEN** docker-compose starts the lango service
- **THEN** the recommended configuration method is `lango config import` with auto-deletion of the source file
- **AND** `LANGO_PASSPHRASE` environment variable SHALL be used for non-interactive passphrase entry
