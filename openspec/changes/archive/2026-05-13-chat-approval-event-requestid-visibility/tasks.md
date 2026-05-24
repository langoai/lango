## 1. Approval Event Traceability

- [x] 1.1 Include compact request-id annotations on approval transcript events.
- [x] 1.2 Add regressions covering the visible request-id annotation.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so approval event visibility mentions compact request-id annotations.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the traceability contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-event-requestid-visibility --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
