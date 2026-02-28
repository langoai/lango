# GoReleaser Release Configuration

## Purpose

Defines the GoReleaser configuration for multi-platform binary builds with standard and extended (KMS) variants, SHA256 checksums, conventional commit changelog, and GitHub Release settings.

## Requirements

### Requirement: GoReleaser v2 configuration
The system SHALL provide a `.goreleaser.yaml` configuration file using GoReleaser v2 schema (`version: 2`) at the project root.

#### Scenario: Configuration schema version
- **WHEN** GoReleaser parses `.goreleaser.yaml`
- **THEN** the configuration SHALL use `version: 2` schema

### Requirement: Standard build variant
The system SHALL define a build named `lango` that compiles `./cmd/lango` with `CGO_ENABLED=1` for linux and darwin on amd64 and arm64 architectures, with ldflags injecting version and build time.

#### Scenario: Standard build targets
- **WHEN** GoReleaser executes the `lango` build
- **THEN** it SHALL produce binaries for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 with `-X main.Version` and `-X main.BuildTime` ldflags

### Requirement: Extended build variant
The system SHALL define a build named `lango-extended` that compiles `./cmd/lango` with `CGO_ENABLED=1` and build tag `kms_all` for the same platform matrix as the standard build.

#### Scenario: Extended build includes KMS tags
- **WHEN** GoReleaser executes the `lango-extended` build
- **THEN** it SHALL compile with `-tags kms_all` producing binaries with AWS/GCP/Azure/PKCS11 KMS support

### Requirement: Archive naming convention
The system SHALL produce tar.gz archives with naming pattern `lango_{{.Version}}_{{.Os}}_{{.Arch}}` for standard and `lango-extended_{{.Version}}_{{.Os}}_{{.Arch}}` for extended builds.

#### Scenario: Standard archive name
- **WHEN** building version v0.3.0 for linux/amd64
- **THEN** the standard archive SHALL be named `lango_0.3.0_linux_amd64.tar.gz`

#### Scenario: Extended archive name
- **WHEN** building version v0.3.0 for darwin/arm64
- **THEN** the extended archive SHALL be named `lango-extended_0.3.0_darwin_arm64.tar.gz`

### Requirement: SHA256 checksums
The system SHALL generate a `checksums.txt` file containing SHA256 hashes for all release artifacts.

#### Scenario: Checksum file generation
- **WHEN** GoReleaser completes all archive builds
- **THEN** it SHALL produce a `checksums.txt` file using SHA256 algorithm

### Requirement: Conventional commit changelog
The system SHALL generate a changelog grouped by conventional commit types: Features (`feat:`), Bug Fixes (`fix:`), Refactoring (`refactor:`), Documentation (`docs:`), and Others.

#### Scenario: Changelog grouping
- **WHEN** GoReleaser generates the changelog
- **THEN** commits SHALL be sorted ascending and grouped by prefix, with `test:`, `chore:`, and `ci:` commits excluded

### Requirement: Release configuration
The system SHALL create GitHub Releases with prerelease auto-detection and non-draft mode, using name template `{{.ProjectName}} v{{.Version}}`.

#### Scenario: Prerelease detection
- **WHEN** a tag like `v0.3.0-rc.1` is pushed
- **THEN** the GitHub Release SHALL be marked as prerelease automatically

#### Scenario: Stable release
- **WHEN** a tag like `v0.3.0` is pushed
- **THEN** the GitHub Release SHALL be created as a stable release (not draft, not prerelease)
