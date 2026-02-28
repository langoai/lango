## 1. Build Job Fix

- [x] 1.1 Replace `goreleaser build --split` with `goreleaser build --single-target --clean --timeout 60m` in the build job
- [x] 1.2 Update build step name from "Build (split)" to "Build (single-target)"

## 2. Release Job Rewrite

- [x] 2.1 Remove Go setup and GoReleaser setup steps from the release job
- [x] 2.2 Add version extraction step (strip `v` prefix from GITHUB_REF_NAME)
- [x] 2.3 Add archive creation step: iterate dist/ dirs, normalize GOAMD64 suffixes, create tar.gz archives
- [x] 2.4 Add SHA256 checksum generation step
- [x] 2.5 Replace `goreleaser continue --merge` with `gh release create --generate-notes`

## 3. Verification

- [x] 3.1 Run `goreleaser check` to validate .goreleaser.yaml remains valid
- [x] 3.2 Verify `--single-target` flag is available in GoReleaser OSS
