## ADDED Requirements

### Requirement: Runtime permission verification
The entrypoint script SHALL verify write permissions on critical directories before starting lango.

#### Scenario: All directories writable
- **WHEN** the entrypoint runs and all critical directories are writable
- **THEN** the script SHALL proceed normally without errors

#### Scenario: Directory not writable due to volume ownership mismatch
- **WHEN** the entrypoint runs and a critical directory is not writable
- **THEN** the script SHALL print an error message to stderr identifying the non-writable directory, the current user, and the UID
- **AND** the script SHALL print a hint suggesting `docker volume rm lango-data`
- **AND** the script SHALL exit with code 1

#### Scenario: Critical directories checked
- **WHEN** the entrypoint runs
- **THEN** it SHALL verify writability of `$HOME/.lango`, `$HOME/.lango/skills`, and `$HOME/bin`

### Requirement: User-writable binary directory
The Docker image SHALL provide a user-writable directory on PATH for installing CLI tools.

#### Scenario: Binary directory exists
- **WHEN** the container starts
- **THEN** `$HOME/bin` SHALL exist and be owned by the lango user
- **AND** `$HOME/bin` SHALL be included in the PATH environment variable

#### Scenario: Agent installs a tool
- **WHEN** an agent downloads or compiles a binary to `$HOME/bin`
- **THEN** the binary SHALL be executable via its name without specifying the full path

### Requirement: Skills subdirectory pre-creation
The Docker image SHALL pre-create the `.lango/skills/` subdirectory with correct ownership.

#### Scenario: Docker volume initialization
- **WHEN** a new named volume is first mounted at `/home/lango/.lango`
- **THEN** the volume SHALL inherit the `skills/` subdirectory with lango:lango ownership

#### Scenario: Entrypoint creates skills directory
- **WHEN** the entrypoint script runs
- **THEN** it SHALL ensure `$HOME/.lango/skills` exists via `mkdir -p`

### Requirement: Optional Go toolchain
The Docker image SHALL support an optional Go toolchain installation via build argument.

#### Scenario: Default build without Go
- **WHEN** the Docker image is built without `--build-arg INSTALL_GO=true`
- **THEN** Go SHALL NOT be installed
- **AND** the image size SHALL not increase

#### Scenario: Build with Go toolchain
- **WHEN** the Docker image is built with `--build-arg INSTALL_GO=true`
- **THEN** Go SHALL be installed at `/usr/local/go`
- **AND** `GOPATH` SHALL be set to `/home/lango/go`
- **AND** both `/home/lango/go/bin` and `/usr/local/go/bin` SHALL be on PATH

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

#### Scenario: Non-root execution
- **WHEN** the container starts
- **THEN** the lango process SHALL run as non-root user
- **AND** WORKDIR SHALL be `/home/lango` (user home directory, writable)
- **AND** the Dockerfile SHALL NOT create a separate `/data` directory
- **AND** `$HOME/.lango/skills/` and `$HOME/bin/` SHALL be pre-created with lango:lango ownership

#### Scenario: Health check
- **WHEN** the container is running
- **THEN** Docker SHALL perform health checks via `lango health` CLI command
- **AND** unhealthy containers SHALL be marked for restart

#### Scenario: Entrypoint script
- **WHEN** the container starts
- **THEN** the system SHALL execute `docker-entrypoint.sh` as the entrypoint
- **AND** the entrypoint SHALL have execute permission set during build
- **AND** the entrypoint SHALL verify write permissions on critical directories
- **AND** the entrypoint SHALL set up passphrase keyfile before starting lango
- **AND** the entrypoint SHALL import config on first run only
- **AND** the entrypoint SHALL `exec lango` to replace itself as PID 1

#### Scenario: Build context optimization
- **WHEN** building the Docker image
- **THEN** `.dockerignore` SHALL exclude `.git`, `.claude`, `openspec/`, and other non-essential files from the build context
