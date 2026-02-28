## 1. GoReleaser Configuration

- [x] 1.1 Create `.goreleaser.yaml` with v2 schema, standard build (`lango`) and extended build (`lango-extended`) targeting linux/darwin × amd64/arm64
- [x] 1.2 Configure archive naming: `lango_{{.Version}}_{{.Os}}_{{.Arch}}` (standard) and `lango-extended_{{.Version}}_{{.Os}}_{{.Arch}}` (extended) as tar.gz
- [x] 1.3 Configure SHA256 checksum generation (`checksums.txt`)
- [x] 1.4 Configure conventional commit changelog with feat/fix/refactor/docs grouping and test/chore/ci exclusion
- [x] 1.5 Configure release settings: prerelease auto, draft false, name template

## 2. Release Workflow

- [x] 2.1 Create `.github/workflows/release.yml` with `push tags: v*` trigger and `contents: write` permission
- [x] 2.2 Configure build job matrix with 4 native runners (ubuntu-latest, ubuntu-24.04-arm, macos-13, macos-14)
- [x] 2.3 Add conditional Linux dependency installation (`libsqlite3-dev`)
- [x] 2.4 Configure `goreleaser build --split --clean` in build job with artifact upload
- [x] 2.5 Configure release job: download artifacts with `merge-multiple: true`, run `goreleaser continue --merge`

## 3. CI Workflow

- [x] 3.1 Create `.github/workflows/ci.yml` with push/PR triggers on `main` and `contents: read` permission
- [x] 3.2 Configure test job matrix (ubuntu-latest + macos-14) with build, test -race -cover, and vet
- [x] 3.3 Add lint job with `golangci-lint-action`
- [x] 3.4 Add goreleaser-check job validating `.goreleaser.yaml`

## 4. Local Development Support

- [x] 4.1 Add `release-dry` Makefile target (`goreleaser build --single-target --snapshot --clean`)
- [x] 4.2 Add `release-check` Makefile target (`goreleaser check`)
- [x] 4.3 Add `dist/` to `.gitignore`
