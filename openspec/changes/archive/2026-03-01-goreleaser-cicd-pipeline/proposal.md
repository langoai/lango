## Why

The project currently relies on manual Makefile-based builds with no automated release pipeline. Creating multi-platform releases requires running builds on each platform individually. GoReleaser with GitHub Actions automates multi-platform binary builds, GitHub Release creation, and changelog generation on tag push, reducing release friction from hours of manual work to a single `git tag && git push`.

## What Changes

- Add `.goreleaser.yaml` with two build variants: standard (default) and extended (`-tags kms_all`)
- Add `release.yml` GitHub Actions workflow using native runner matrix + `--split`/`--merge` strategy for CGO-dependent cross-platform builds (linux/darwin × amd64/arm64)
- Add `ci.yml` GitHub Actions workflow for PR/push validation (build, test, vet, lint, goreleaser check)
- Add `release-dry` and `release-check` Makefile targets for local testing
- Add `dist/` to `.gitignore`

## Capabilities

### New Capabilities
- `goreleaser-release`: GoReleaser configuration for multi-platform binary builds with standard/extended variants, SHA256 checksums, and conventional commit changelog
- `release-workflow`: GitHub Actions release pipeline using native runner matrix with split/merge strategy for CGO cross-compilation
- `ci-workflow`: GitHub Actions CI pipeline for automated build, test, lint, and config validation on PR/push

### Modified Capabilities

## Impact

- **New files**: `.goreleaser.yaml`, `.github/workflows/release.yml`, `.github/workflows/ci.yml`
- **Modified files**: `Makefile` (new targets), `.gitignore` (dist/ exclusion)
- **Dependencies**: GoReleaser v2 (installed via GitHub Action), golangci-lint (via GitHub Action)
- **Systems**: GitHub Actions runners (ubuntu-latest, ubuntu-24.04-arm, macos-13, macos-14)
- **Secrets**: `GITHUB_TOKEN` (automatically provided by GitHub Actions)
