## 1. Approval Event Summary

- [x] 1.1 Include a compact request summary preview on approval transcript events when a summary exists.
- [x] 1.2 Keep the summary single-line-safe and concise.
- [x] 1.3 Add regressions for the visible summary preview.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so approval-event visibility mentions the summary preview.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the summary-preview contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-event-summary-visibility --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
