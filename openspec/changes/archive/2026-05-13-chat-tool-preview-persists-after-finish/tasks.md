## 1. Lifecycle Continuity

- [x] 1.1 Preserve the compact param preview when a running tool row transitions to success.
- [x] 1.2 Preserve the compact param preview when a running tool row transitions to error.
- [x] 1.3 Add regressions for the persisted preview.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so tool lifecycle visibility mentions that the compact param preview survives completion.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the persisted-preview contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-tool-preview-persists-after-finish --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
