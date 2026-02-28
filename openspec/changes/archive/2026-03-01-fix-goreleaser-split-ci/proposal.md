## Why

The `release.yml` workflow uses `goreleaser build --split` and `goreleaser continue --merge`, which are **GoReleaser Pro-only** flags. The `goreleaser/goreleaser-action@v7` installs the OSS version, causing CI failures with `unknown flag: --split`.

## What Changes

- Replace `goreleaser build --split` with `goreleaser build --single-target` (OSS-compatible) in the build job
- Replace `goreleaser continue --merge` with manual archive creation + `gh release create` in the release job
- Remove unnecessary Go setup step from the release job
- Add version extraction, tar.gz archive creation, SHA256 checksum generation, and GitHub Release creation steps

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `release-workflow`: Replace Pro-only split/merge strategy with OSS-compatible single-target build + manual archive/release pipeline

## Impact

- **File**: `.github/workflows/release.yml` — build and release jobs restructured
- **No code changes**: `.goreleaser.yaml` unchanged; build settings (ldflags, CGO_ENABLED, tags) still used by `--single-target`
- **Release artifacts**: Same naming convention maintained (`lango_{VERSION}_{os}_{arch}.tar.gz`, `checksums.txt`)
