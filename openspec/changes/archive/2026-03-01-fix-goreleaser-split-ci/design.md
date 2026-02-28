## Context

The release workflow uses `goreleaser build --split` and `goreleaser continue --merge`, which are GoReleaser Pro-only features. The `goreleaser/goreleaser-action@v7` installs the OSS version, causing `unknown flag: --split` errors. This project requires CGO_ENABLED=1 (mattn/go-sqlite3), making cross-compilation impossible and requiring native per-platform builds.

## Goals / Non-Goals

**Goals:**
- Make the release workflow work with GoReleaser OSS (free tier)
- Maintain the same release artifact naming convention
- Keep the native runner matrix strategy for CGO cross-platform builds
- Produce identical release output (tar.gz archives + checksums)

**Non-Goals:**
- Migrating to GoReleaser Pro
- Changing the `.goreleaser.yaml` build configuration
- Modifying the archive naming convention
- Adding new platforms or architectures

## Decisions

### Decision 1: Use `--single-target` instead of `--split`
`goreleaser build --single-target` is an OSS-compatible flag that builds only for the current GOOS/GOARCH. Combined with the existing matrix of native runners, this produces the same result as `--split` — each runner builds only its own platform binaries.

**Alternative considered**: Running `go build` directly — rejected because it would duplicate ldflags/tags/env logic already defined in `.goreleaser.yaml`.

### Decision 2: Manual archive + release instead of `--merge`
Since `goreleaser continue --merge` is Pro-only, the release job manually:
1. Finds built binaries in downloaded `dist/` directories
2. Normalizes GOAMD64 suffixes (e.g., `_v1`) from directory names
3. Creates tar.gz archives with the same naming convention
4. Generates SHA256 checksums
5. Uses `gh release create --generate-notes` for the GitHub Release

**Alternative considered**: Using a separate release tool (e.g., `nfpm`) — rejected as overkill for tar.gz archives.

### Decision 3: Remove Go setup from release job
The release job no longer runs GoReleaser, so Go is not needed. Only shell tools (`tar`, `sha256sum`, `gh`) are required.

## Risks / Trade-offs

- [Risk] GOAMD64 suffix normalization via `sed 's/_v[0-9]*$//'` could match unintended patterns → Mitigation: Pattern is specific to trailing `_v` + digits, matching GoReleaser's known convention
- [Risk] LICENSE/README.md may not exist in some builds → Mitigation: Fallback tar command without docs files
- [Risk] `sha256sum` not available on all runners → Mitigation: Release job runs on ubuntu-latest where it's always available
- [Trade-off] Changelog grouping from `.goreleaser.yaml` is no longer used; `--generate-notes` uses GitHub's default format → Acceptable since GitHub's auto-generated notes are sufficient
