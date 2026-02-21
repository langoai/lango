## Purpose

Define the Docker container configuration, compose orchestration, and secure secrets-based deployment model for Lango.
## Requirements
### Requirement: Docker Container Configuration
The system SHALL provide a Dockerfile optimized for production deployment.

#### Scenario: Multi-stage build
- **WHEN** building the Docker image
- **THEN** the system SHALL use a multi-stage build
- **AND** the builder stage SHALL compile with CGO_ENABLED=1
- **AND** the builder stage SHALL use `--no-install-recommends` for apt packages
- **AND** the runtime stage SHALL use debian:bookworm-slim

#### Scenario: Browser always included
- **WHEN** building the Docker image
- **THEN** the runtime image SHALL always include Chromium browser via `--no-install-recommends`
- **AND** no build arguments SHALL control Chromium inclusion

#### Scenario: No curl dependency
- **WHEN** the Docker image is built
- **THEN** the runtime image SHALL NOT include curl
- **AND** health checks SHALL use `lango health` CLI command instead

#### Scenario: Non-root execution
- **WHEN** the container starts
- **THEN** the lango process SHALL run as non-root user
- **AND** WORKDIR SHALL be `/home/lango` (user home directory, writable)
- **AND** the Dockerfile SHALL NOT create a separate `/data` directory

#### Scenario: Health check
- **WHEN** the container is running
- **THEN** Docker SHALL perform health checks via `lango health` CLI command
- **AND** unhealthy containers SHALL be marked for restart

#### Scenario: Entrypoint script
- **WHEN** the container starts
- **THEN** the system SHALL execute `docker-entrypoint.sh` as the entrypoint
- **AND** the entrypoint SHALL have execute permission set during build
- **AND** the entrypoint SHALL set up passphrase keyfile before starting lango
- **AND** the entrypoint SHALL import config on first run only
- **AND** the entrypoint SHALL `exec lango` to replace itself as PID 1

#### Scenario: Build context optimization
- **WHEN** building the Docker image
- **THEN** `.dockerignore` SHALL exclude `.git`, `.claude`, `openspec/`, and other non-essential files from the build context

### Requirement: Docker Compose Orchestration
The system SHALL provide a docker-compose.yml with a single lango service.

#### Scenario: Service definition
- **WHEN** running `docker compose up -d`
- **THEN** the lango service SHALL expose port 18789
- **AND** volumes SHALL persist data to lango-data volume

#### Scenario: Single service deployment
- **WHEN** running `docker compose up -d`
- **THEN** only the lango service SHALL start
- **AND** no profiles SHALL be required
- **AND** the image SHALL include Chromium for browser automation

#### Scenario: Configuration via import
- **WHEN** docker-compose starts the lango service
- **THEN** the recommended configuration method is `lango config import` with auto-deletion of the source file
- **AND** `LANGO_PASSPHRASE` environment variable SHALL be used for non-interactive passphrase entry

### Requirement: Data Persistence
The system SHALL persist data across container restarts.

#### Scenario: SQLite database persistence
- **WHEN** the container restarts
- **THEN** the SQLite database at /home/lango/.lango SHALL be preserved
- **AND** session history SHALL not be lost

#### Scenario: Volume mount
- **WHEN** docker-compose is used
- **THEN** a named volume (lango-data) SHALL be mounted at /home/lango/.lango

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

### Requirement: Makefile Docker targets
The Makefile SHALL provide targets for managing Docker containers.

#### Scenario: Start containers
- **WHEN** running `make docker-up`
- **THEN** the system SHALL execute `docker compose up -d`

#### Scenario: Stop containers
- **WHEN** running `make docker-down`
- **THEN** the system SHALL execute `docker compose down`

#### Scenario: Tail logs
- **WHEN** running `make docker-logs`
- **THEN** the system SHALL follow container logs via `docker compose logs -f`

### Requirement: Makefile Docker build
The Makefile SHALL provide targets for building Docker images.

#### Scenario: Build with latest tag
- **WHEN** running `make docker-build`
- **THEN** the system SHALL tag the image with both the version tag and `latest`

#### Scenario: Push to registry
- **WHEN** running `make docker-push REGISTRY=my.registry.io`
- **THEN** the system SHALL tag and push both version and latest tags to the specified registry
- **WHEN** running `make docker-push` without REGISTRY set
- **THEN** the system SHALL fail with an error message

### Requirement: Presidio analyzer Docker service
The docker-compose.yml SHALL include a presidio-analyzer service using the mcr.microsoft.com/presidio-analyzer:latest image, exposed on port 5002, under the "presidio" profile.

#### Scenario: Profile-based activation
- **WHEN** user runs `docker compose --profile presidio up`
- **THEN** the presidio-analyzer container SHALL start alongside the main lango service

#### Scenario: Default compose up
- **WHEN** user runs `docker compose up` without the presidio profile
- **THEN** the presidio-analyzer container SHALL NOT start

### Requirement: Docker config example includes Presidio service fields
The Docker deployment example config.json SHALL include the `presidio` block within `security.interceptor` so users deploying with `--profile presidio` have the correct default configuration.

#### Scenario: User deploys with Presidio profile
- **WHEN** a user runs `docker compose --profile presidio up` with the example config
- **THEN** the config.json already contains `presidio.enabled: false`, `presidio.url: "http://localhost:5002"`, `presidio.scoreThreshold: 0.7`, and `presidio.language: "en"`
- **THEN** the user only needs to set `presidio.enabled: true` to activate Presidio detection

