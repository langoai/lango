## MODIFIED Requirements

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
