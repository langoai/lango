## 1. UX Fix

- [x] 1.1 Hide Tools navigation help when there is no navigable category list.
- [x] 1.2 Hide Sessions navigation help when there are no navigable session rows.
- [x] 1.3 Add regression coverage for both empty-state help contracts.

## 2. Docs And Spec Sync

- [x] 2.1 Update the Tools and Sessions page specs to require hiding inert navigation help.
- [x] 2.2 Update cockpit feature docs to describe that navigation hints appear only when rows exist.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
