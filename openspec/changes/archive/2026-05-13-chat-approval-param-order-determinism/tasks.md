## 1. Deterministic Param Rendering

- [x] 1.1 Render approval request params in stable key order in the fallback approval banner.
- [x] 1.2 Render approval request params in stable key order in the fullscreen approval dialog.
- [x] 1.3 Add regressions covering ordered param output.

## 2. Downstream Sync

- [x] 2.1 Update public cockpit docs to mention that approval params render in stable key order.
- [x] 2.2 Update the `tui-chat-rendering` OpenSpec delta with the deterministic param-order contract.

## 3. Verification

- [x] 3.1 Run `go test ./internal/cli/chat -count=1`.
- [x] 3.2 Run `go build ./...`.
- [x] 3.3 Run `go test ./...`.
- [x] 3.4 Run `openspec validate chat-approval-param-order-determinism --strict`.
- [x] 3.5 Validate and archive the OpenSpec change.
