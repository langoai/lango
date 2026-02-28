## ADDED Requirements

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
Each matrix runner SHALL execute `goreleaser build --split --clean` to produce binaries only for its native platform.

#### Scenario: Split build produces platform-specific artifacts
- **WHEN** `goreleaser build --split` runs on `macos-14`
- **THEN** it SHALL produce darwin/arm64 binaries only and upload them as artifacts

### Requirement: Merge and release job
A separate `release` job SHALL download all build artifacts, merge them into `dist/`, and run `goreleaser continue --merge` to create the GitHub Release.

#### Scenario: Artifact merge and release creation
- **WHEN** all 4 build jobs complete successfully
- **THEN** the release job SHALL download artifacts with `merge-multiple: true`, run `goreleaser continue --merge`, and create a GitHub Release with all 8 archives + checksums

### Requirement: Write permissions for release
The workflow SHALL request `contents: write` permission for creating GitHub Releases.

#### Scenario: Permission scope
- **WHEN** the release workflow runs
- **THEN** it SHALL have `contents: write` permission to create releases and upload assets
