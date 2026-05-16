## 1. UX Fix

- [x] 1.1 Hide Mission Control `↑/↓` help when the focused lane has fewer than two navigable rows.
- [x] 1.2 Add regression coverage for single-row and multi-row help states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require navigation help only when another focused-lane row exists.
- [x] 2.2 Update cockpit feature docs to describe the same help contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
