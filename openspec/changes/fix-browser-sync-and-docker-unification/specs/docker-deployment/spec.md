## MODIFIED Requirements

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

## REMOVED Requirements

### Requirement: Slim deployment (default profile)
**Reason**: Single image always includes Chromium. No need for slim variant without browser.
**Migration**: Use `docker compose up -d` without any profile flags.

### Requirement: Built-in browser deployment (browser profile)
**Reason**: Merged into default single image. Chromium is always included.
**Migration**: Use `docker compose up -d` without any profile flags.

### Requirement: Sidecar browser deployment (browser-sidecar profile)
**Reason**: Remote browser support removed. Single image includes Chromium directly.
**Migration**: Use `docker compose up -d`. Remove Chrome sidecar containers.

### Requirement: Makefile browser build variant
**Reason**: No `WITH_BROWSER` build arg. Single image always includes Chromium.
**Migration**: Use `make docker-build` for the unified image.

### Requirement: Makefile browser/sidecar compose targets
**Reason**: No compose profiles. Single `docker compose up -d` command.
**Migration**: Use `make docker-up` and `make docker-down`.
