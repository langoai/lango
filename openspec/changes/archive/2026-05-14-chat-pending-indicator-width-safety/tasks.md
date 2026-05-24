## 1. Pending Indicator Hardening

- [x] 1.1 Make the submit-to-first-event pending indicator width-aware.
- [x] 1.2 Use the width-aware renderer in both view assembly and layout measurement.
- [x] 1.3 Add regressions for narrow-width rendering.

## 2. Spec Sync

- [x] 2.1 Record the pending-indicator width-safety contract in the `tui-chat-rendering` delta.
- [x] 2.2 Update downstream feature docs to match the runtime behavior.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-pending-indicator-width-safety --strict`.
- [x] 3.5 Sync/archive the OpenSpec change and run `openspec validate --specs`.
