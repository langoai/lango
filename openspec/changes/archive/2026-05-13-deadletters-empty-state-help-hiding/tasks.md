## 1. UX Fix

- [x] 1.1 Hide Dead Letters `↑/↓` help when no backlog rows exist.
- [x] 1.2 Add regression coverage for empty and populated help states.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require hiding inert row-navigation help in empty states.
- [x] 2.2 Update cockpit feature docs to describe that row-navigation hints appear only when backlog rows exist.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
