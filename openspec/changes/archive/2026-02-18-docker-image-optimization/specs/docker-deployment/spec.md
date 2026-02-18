## MODIFIED Requirements

### Requirement: Docker Container Configuration
The system SHALL provide a Dockerfile optimized for production deployment.

#### Scenario: Multi-stage build
- **WHEN** building the Docker image
- **THEN** the system SHALL use a multi-stage build
- **AND** the builder stage SHALL compile with CGO_ENABLED=1
- **AND** the builder stage SHALL use `--no-install-recommends` for apt packages
- **AND** the runtime stage SHALL use debian:bookworm-slim

#### Scenario: Conditional browser tool support
- **WHEN** Docker image is built with `WITH_BROWSER=false` (default)
- **THEN** the runtime image SHALL NOT include Chromium browser
- **AND** the resulting image SHALL be approximately 200MB
- **WHEN** Docker image is built with `--build-arg WITH_BROWSER=true`
- **THEN** the runtime image SHALL include Chromium browser via `--no-install-recommends`

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
The system SHALL provide a docker-compose.yml with deployment profiles for different browser configurations.

#### Scenario: Slim deployment (default profile)
- **WHEN** running `docker compose up` or `docker compose --profile default up`
- **THEN** the lango service SHALL start without Chromium
- **AND** the image SHALL be approximately 200MB

#### Scenario: Built-in browser deployment (browser profile)
- **WHEN** running `docker compose --profile browser up`
- **THEN** the lango-browser service SHALL start with Chromium included in the image

#### Scenario: Sidecar browser deployment (browser-sidecar profile)
- **WHEN** running `docker compose --profile browser-sidecar up`
- **THEN** the lango-sidecar service SHALL start without Chromium
- **AND** a separate `chromedp/headless-shell` container SHALL run alongside
- **AND** lango SHALL connect to Chrome via `ROD_BROWSER_WS=ws://chrome:9222`
- **AND** the Chrome container SHALL have a memory limit of 512MB

#### Scenario: Service definition
- **WHEN** running any docker compose profile
- **THEN** the lango service SHALL expose port 18789
- **AND** volumes SHALL persist data to lango-data volume
