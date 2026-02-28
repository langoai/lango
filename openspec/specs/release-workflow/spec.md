# Release Workflow

## Purpose

Defines the GitHub Actions release pipeline that uses a native runner matrix with single-target builds and manual archive/release for CGO-dependent cross-platform binary builds on tag push.

## Requirements

### Requirement: Tag-triggered release workflow
The system SHALL provide a GitHub Actions workflow at `.github/workflows/release.yml` that triggers on push of tags matching `v*`.

#### Scenario: Workflow trigger
- **WHEN** a tag `v0.3.0` is pushed to the repository
- **THEN** the release workflow SHALL start automatically

#### Scenario: Non-tag push ignored
- **WHEN** a commit is pushed to `main` without a tag
- **THEN** the release workflow SHALL NOT trigger

### Requirement: Native runner matrix build
The build job SHALL use a strategy matrix with 4 native runners: `ubuntu-latest` (linux/amd64), `ubuntu-24.04-arm` (linux/arm64), `macos-13` (darwin/amd64), `macos-14` (darwin/arm64).

#### Scenario: Matrix runner assignment
- **WHEN** the build job starts
- **THEN** it SHALL spawn 4 parallel jobs, one per runner in the matrix

### Requirement: Linux dependency installation
The workflow SHALL install `libsqlite3-dev` on Linux runners before building.

#### Scenario: Linux build dependencies
- **WHEN** the build job runs on a Linux runner
- **THEN** it SHALL run `apt-get install -y libsqlite3-dev` before GoReleaser

#### Scenario: macOS skips dependency install
- **WHEN** the build job runs on a macOS runner
- **THEN** it SHALL NOT run apt-get (macOS uses system frameworks)

### Requirement: Split build execution
Each matrix runner SHALL execute `goreleaser build --single-target --clean --timeout 60m` to produce binaries only for its native platform. The `--single-target` flag uses the GOOS and GOARCH environment variables set by the matrix to determine the build target.

#### Scenario: Single-target build produces platform-specific artifacts
- **WHEN** `goreleaser build --single-target` runs on `macos-14` with GOOS=darwin GOARCH=arm64
- **THEN** it SHALL produce darwin/arm64 binaries for both build IDs (lango and lango-extended) and upload the dist/ directory as artifacts

### Requirement: Merge and release job
A separate `release` job SHALL download all build artifacts and create a GitHub Release with manually assembled archives and checksums. The job SHALL NOT require Go or GoReleaser installation.

#### Scenario: Manual archive creation and release
- **WHEN** all 4 build jobs complete successfully
- **THEN** the release job SHALL:
  1. Download artifacts with `merge-multiple: true`
  2. Extract version from the git tag (strip `v` prefix)
  3. Find built binaries in dist/ subdirectories
  4. Normalize directory names by stripping GOAMD64 suffixes (e.g., `_v1`)
  5. Create tar.gz archives named `{build_id}_{VERSION}_{os}_{arch}.tar.gz`
  6. Generate SHA256 checksums in `checksums.txt`
  7. Create a GitHub Release using `gh release create` with `--generate-notes`

#### Scenario: Archive naming convention
- **WHEN** archives are created for version 0.3.0
- **THEN** they SHALL follow the naming pattern `lango_0.3.0_linux_amd64.tar.gz` and `lango-extended_0.3.0_linux_amd64.tar.gz`

### Requirement: Write permissions for release
The workflow SHALL request `contents: write` permission for creating GitHub Releases.

#### Scenario: Permission scope
- **WHEN** the release workflow runs
- **THEN** it SHALL have `contents: write` permission to create releases and upload assets
