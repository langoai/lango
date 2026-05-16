## 1. Markdown Fallback Hardening

- [x] 1.1 Make markdown rendering fall back to plain text on renderer errors.
- [x] 1.2 Make markdown rendering recover to plain text on renderer panics.
- [x] 1.3 Add regressions for both fallback paths.

## 2. Verification

- [x] 2.1 Run `go test ./internal/cli/chat ./internal/cli/cockpit/pages -count=1`.
- [x] 2.2 Run `go build ./...`.
- [x] 2.3 Run `go test ./...`.

## 3. Spec Sync

- [x] 3.1 Update `tui-chat-rendering` spec coverage for markdown fallback behavior.
- [x] 3.2 Validate and archive the OpenSpec change.
