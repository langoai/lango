## 1. Recovery Row Hardening

- [x] 1.1 Make `renderRecoveryBlock()` clamp to the requested width.
- [x] 1.2 Add regressions for narrow-width recovery rendering.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so recovery event rendering mentions the compact row staying one-line on narrow terminals.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the width-safety contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-recovery-block-width-safety --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
