## 1. UX Copy

- [x] 1.1 Update the default Mission Control composer hint to neutral request wording.
- [x] 1.2 Update the workbench footer/helper copy that reused the same phrase.
- [x] 1.3 Update regressions that pin the first-screen composer hint.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit docs to describe the current first-screen composer hint.
- [x] 2.2 Extend downstream docs requirements so the hint wording stays aligned.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages ./internal/cli/workbench -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
