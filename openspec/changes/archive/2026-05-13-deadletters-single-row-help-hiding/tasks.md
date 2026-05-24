## 1. UX Fix

- [x] 1.1 Hide Dead Letters `↑/↓` help when fewer than two backlog rows exist.
- [x] 1.2 Add regression coverage for the single-row help state.

## 2. Docs And Spec Sync

- [x] 2.1 Update the cockpit-pages spec to require row-navigation help only when another row exists.
- [x] 2.2 Update cockpit feature docs to describe the same contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
