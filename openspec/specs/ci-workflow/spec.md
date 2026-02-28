# CI Workflow

## Purpose

Defines the GitHub Actions CI pipeline for automated build, test, lint, and GoReleaser config validation on pull requests and pushes to main.

## Requirements

### Requirement: CI workflow triggers
The system SHALL provide a GitHub Actions workflow at `.github/workflows/ci.yml` that triggers on push to `main` and pull requests targeting `main`.

#### Scenario: PR trigger
- **WHEN** a pull request is opened targeting `main`
- **THEN** the CI workflow SHALL start automatically

#### Scenario: Main branch push trigger
- **WHEN** a commit is pushed to `main`
- **THEN** the CI workflow SHALL start automatically

### Requirement: Multi-platform test job
The test job SHALL run on both Linux (`ubuntu-latest`) and macOS (`macos-14`) runners with CGO enabled.

#### Scenario: Test matrix execution
- **WHEN** the test job starts
- **THEN** it SHALL run `go build ./...`, `go test -race -cover ./...`, and `go vet ./...` on both platforms

### Requirement: Linux test dependencies
The test job SHALL install `libsqlite3-dev` on Linux runners.

#### Scenario: Linux CI dependencies
- **WHEN** the test job runs on Linux
- **THEN** it SHALL install `libsqlite3-dev` via apt-get before building

### Requirement: Lint job
The CI workflow SHALL include a lint job running `golangci-lint` on Linux using the official `golangci-lint-action`.

#### Scenario: Lint execution
- **WHEN** the lint job runs
- **THEN** it SHALL execute golangci-lint with the latest version

### Requirement: GoReleaser config validation job
The CI workflow SHALL include a job that validates `.goreleaser.yaml` by running `goreleaser check`.

#### Scenario: Config validation
- **WHEN** the goreleaser-check job runs
- **THEN** it SHALL execute `goreleaser check` and fail if the configuration is invalid

### Requirement: Read-only permissions
The CI workflow SHALL request only `contents: read` permission.

#### Scenario: CI permission scope
- **WHEN** the CI workflow runs
- **THEN** it SHALL operate with `contents: read` permission only (no write access)
