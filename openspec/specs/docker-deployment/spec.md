## Purpose

Define the Docker container configuration, compose orchestration, and secure secrets-based deployment model for Lango.
## Requirements
### Requirement: Docker Container Configuration
The system SHALL provide a Dockerfile optimized for production deployment.

#### Scenario: Multi-stage build
- **WHEN** building the Docker image
- **THEN** the system SHALL use a multi-stage build
- **AND** the builder stage SHALL compile with CGO_ENABLED=1
- **AND** the runtime stage SHALL use debian:bookworm-slim

#### Scenario: Browser tool support
- **WHEN** Docker image is built
- **THEN** the runtime image SHALL include Chromium browser
- **AND** go-rod SHALL auto-detect the system Chromium via `launcher.LookPath()`

#### Scenario: Non-root execution
- **WHEN** the container starts
- **THEN** the lango process SHALL run as non-root user
- **AND** WORKDIR SHALL be `/home/lango` (user home directory, writable)

#### Scenario: Health check
- **WHEN** the container is running
- **THEN** Docker SHALL perform health checks via HTTP endpoint
- **AND** unhealthy containers SHALL be marked for restart

#### Scenario: Entrypoint script
- **WHEN** the container starts
- **THEN** the system SHALL execute `docker-entrypoint.sh` as the entrypoint
- **AND** the entrypoint SHALL have execute permission set during build
- **AND** the entrypoint SHALL set up passphrase keyfile before starting lango
- **AND** the entrypoint SHALL import config on first run only
- **AND** the entrypoint SHALL `exec lango` to replace itself as PID 1

### Requirement: Docker Compose Orchestration
The system SHALL provide a docker-compose.yml for simplified deployment. The README documentation SHALL describe the import→delete configuration pattern instead of read-only JSON mounting.

#### Scenario: Service definition
- **WHEN** running `docker-compose up`
- **THEN** the lango service SHALL start on port 18789
- **AND** volumes SHALL persist data to lango-data volume

#### Scenario: Optional prompts volume mount
- **WHEN** docker-compose.yml is inspected
- **THEN** it SHALL contain a commented-out volume mount for `./prompts:/usr/share/lango/prompts` to allow runtime prompt customization
- **AND** the default behavior (using embedded prompts) SHALL be unchanged when the comment is not removed

#### Scenario: Configuration via import
- **WHEN** docker-compose starts the lango service
- **THEN** the recommended configuration method is `lango config import` with auto-deletion of the source file
- **AND** `LANGO_PASSPHRASE` environment variable SHALL be used for non-interactive passphrase entry

### Requirement: Data Persistence
The system SHALL persist data across container restarts.

#### Scenario: SQLite database persistence
- **WHEN** the container restarts
- **THEN** the SQLite database at /data SHALL be preserved
- **AND** session history SHALL not be lost

#### Scenario: Volume mount
- **WHEN** docker-compose is used
- **THEN** a named volume (lango-data) SHALL be mounted at /data

### Requirement: Secure Entrypoint Config Bootstrap
The system SHALL provide a `docker-entrypoint.sh` script that securely bootstraps configuration before starting lango.

#### Scenario: Passphrase keyfile setup
- **WHEN** the entrypoint script runs
- **AND** a passphrase secret exists at `/run/secrets/lango_passphrase`
- **THEN** the script SHALL copy the passphrase to `~/.lango/keyfile`
- **AND** the keyfile SHALL have permissions 0600
- **AND** the keyfile path SHALL be blocked by the agent's filesystem tool

#### Scenario: First-run config import
- **WHEN** the entrypoint script runs
- **AND** a config secret exists at `/run/secrets/lango_config`
- **AND** no profile with the configured name exists
- **THEN** the script SHALL copy the config to `/tmp/lango-import.json`
- **AND** the script SHALL run `lango config import` with the temp file
- **AND** the temp file SHALL be auto-deleted after import

#### Scenario: Subsequent restart (idempotent)
- **WHEN** the entrypoint script runs
- **AND** the profile already exists in the database
- **THEN** the script SHALL skip the import step
- **AND** the script SHALL proceed to start lango normally

#### Scenario: Configurable secret paths
- **WHEN** the user sets `LANGO_CONFIG_FILE` or `LANGO_PASSPHRASE_FILE` environment variables
- **THEN** the entrypoint SHALL use the specified paths instead of the default Docker secret paths

### Requirement: Headless configuration via import
The system SHALL document a headless configuration pattern for Docker/CI environments where interactive setup is unavailable.

#### Scenario: Docker import workflow
- **WHEN** a Docker container needs configuration without interactive TUI
- **THEN** the user SHALL prepare a JSON config file, COPY it into the container, and run `lango config import config.json --profile production`
- **AND** the source JSON file is automatically deleted after import for security

#### Scenario: Non-interactive passphrase
- **WHEN** running in a headless environment without a terminal
- **THEN** the user SHALL set the `LANGO_PASSPHRASE` environment variable for non-interactive passphrase entry

