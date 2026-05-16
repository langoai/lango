## 1. UX Fix

- [x] 1.1 Render Sessions help bindings with `↑/k` and `↓/j` labels.
- [x] 1.2 Render Tools help bindings with `↑/k` and `↓/j` labels.
- [x] 1.3 Add regressions for the rendered help labels.

## 2. Docs And Spec Sync

- [x] 2.1 Update cockpit page specs to require arrow-style vertical navigation hints.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/cockpit/pages -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
