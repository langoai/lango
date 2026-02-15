## MODIFIED Requirements

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
- **AND** the ROD_BROWSER environment variable SHALL be set

#### Scenario: Non-root execution
- **WHEN** the container starts
- **THEN** the lango process SHALL run as non-root user

#### Scenario: Health check
- **WHEN** the container is running
- **THEN** Docker SHALL perform health checks via HTTP endpoint
- **AND** unhealthy containers SHALL be marked for restart

#### Scenario: Entrypoint script
- **WHEN** the container starts
- **THEN** the system SHALL execute `docker-entrypoint.sh` as the entrypoint
- **AND** the entrypoint SHALL set up passphrase keyfile before starting lango
- **AND** the entrypoint SHALL import config on first run only
- **AND** the entrypoint SHALL `exec lango` to replace itself as PID 1

### Requirement: Docker Compose Orchestration
The system SHALL provide a docker-compose.yml for simplified deployment.

#### Scenario: Service definition
- **WHEN** running `docker-compose up`
- **THEN** the lango service SHALL start on port 18789
- **AND** volumes SHALL persist data to lango-data volume

#### Scenario: Secret injection via Docker secrets
- **WHEN** docker-compose starts the lango service
- **THEN** the system SHALL mount `config.json` as Docker secret `lango_config`
- **AND** the system SHALL mount `passphrase.txt` as Docker secret `lango_passphrase`
- **AND** secrets SHALL be available at `/run/secrets/` (tmpfs-backed)
- **AND** no API keys or passphrases SHALL be passed as environment variables

#### Scenario: Profile configuration
- **WHEN** docker-compose starts the lango service
- **THEN** the `LANGO_PROFILE` environment variable SHALL specify the profile name
- **AND** the default profile name SHALL be `default`

## ADDED Requirements

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
