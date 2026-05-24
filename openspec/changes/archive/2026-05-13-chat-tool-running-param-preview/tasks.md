## 1. Running Tool Preview

- [x] 1.1 Preserve a compact param preview on running tool transcript rows.
- [x] 1.2 Keep the preview ordering deterministic.
- [x] 1.3 Add regressions for the preview output.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so tool lifecycle visibility mentions the running-row param preview.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the preview contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-tool-running-param-preview --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
