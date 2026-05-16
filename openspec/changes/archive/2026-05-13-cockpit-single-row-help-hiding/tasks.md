## 1. UX Fix

- [x] 1.1 Hide Tools `↑/↓` help when fewer than two categories exist.
- [x] 1.2 Hide Sessions `↑/↓` help when fewer than two session rows exist.
- [x] 1.3 Add regression coverage for the single-row/single-category states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the Tools and Sessions specs to require navigation help only when there is another row to move to.
- [x] 2.2 Update cockpit docs to describe the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
