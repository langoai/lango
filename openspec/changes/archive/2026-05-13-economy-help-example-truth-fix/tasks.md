## 1. CLI Help Fix

- [x] 1.1 Replace stale top-level `lango economy` examples with commands that actually exist.
- [x] 1.2 Add a CLI regression covering the corrected help examples.

## 2. Docs And Spec Sync

- [x] 2.1 Update the public economy CLI docs to use the same current-surface examples.
- [x] 2.2 Update the economy CLI spec so top-level help examples stay aligned with the real subcommand tree.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/economy -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Validate and archive the OpenSpec change.
