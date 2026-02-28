## ADDED Requirements

### Requirement: Docker build accepts version build arguments
The Dockerfile SHALL declare `VERSION` and `BUILD_TIME` as `ARG` directives with default values `dev` and `unknown` respectively.

#### Scenario: Build with explicit version arguments
- **WHEN** `docker build --build-arg VERSION=1.0.0 --build-arg BUILD_TIME=2026-03-01T00:00:00Z -t lango .` is executed
- **THEN** the resulting binary SHALL report `lango 1.0.0 (built 2026-03-01T00:00:00Z)` when running `lango version`

#### Scenario: Build without version arguments
- **WHEN** `docker build -t lango .` is executed without `--build-arg`
- **THEN** the resulting binary SHALL report `lango dev (built unknown)` when running `lango version`

### Requirement: Ldflags inject version into Go binary
The `go build` command in the Dockerfile SHALL include `-X main.Version=${VERSION}` and `-X main.BuildTime=${BUILD_TIME}` in the `-ldflags` string, matching the Makefile's injection pattern.

#### Scenario: Ldflags format matches Makefile
- **WHEN** the Dockerfile's `go build` command is inspected
- **THEN** it SHALL contain `-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}` in the ldflags, alongside the existing `-s -w` flags
