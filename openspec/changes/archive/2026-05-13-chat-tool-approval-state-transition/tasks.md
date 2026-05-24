## 1. Tool Approval State Transitions

- [x] 1.1 Move the latest matching running tool row into `awaiting approval` when an approval request arrives.
- [x] 1.2 Restore that row to `running` on approval and mark it `canceled` on denial.
- [x] 1.3 Keep the compact param preview visible through those transitions.
- [x] 1.4 Add regressions for the lifecycle state changes.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs so tool lifecycle visibility mentions the approval/canceled transitions.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the tool-row approval-state contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-tool-approval-state-transition --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
