## 1. Core Types

- [x] 1.1 Add `ApprovalResponse` struct with `Approved` and `AlwaysAllow` fields to `internal/approval/approval.go`
- [x] 1.2 Change `Provider` interface `RequestApproval` return type from `(bool, error)` to `(ApprovalResponse, error)`
- [x] 1.3 Create `internal/approval/grant.go` with `GrantStore` (NewGrantStore, Grant, IsGranted, Revoke, RevokeSession)
- [x] 1.4 Write unit tests for `GrantStore` in `internal/approval/grant_test.go`

## 2. Provider Updates

- [x] 2.1 Update `CompositeProvider.RequestApproval` return type in `internal/approval/composite.go`
- [x] 2.2 Update `HeadlessProvider.RequestApproval` to return `ApprovalResponse{Approved: true}` in `internal/approval/headless.go`
- [x] 2.3 Update `TTYProvider.RequestApproval` with `[y/a/N]` prompt and `AlwaysAllow` support in `internal/approval/tty.go`
- [x] 2.4 Update `GatewayApprover` interface and `GatewayProvider.RequestApproval` return type in `internal/approval/gateway.go`

## 3. Gateway Server

- [x] 3.1 Change `pendingApprovals` map type from `chan bool` to `chan approval.ApprovalResponse` in `internal/gateway/server.go`
- [x] 3.2 Update `RequestApproval` method return type to `(approval.ApprovalResponse, error)`
- [x] 3.3 Add `alwaysAllow` JSON field parsing in `handleApprovalResponse`

## 4. Channel UI

- [x] 4.1 Add "Always Allow" button (Row 2) and `always:` callback handling to Telegram provider
- [x] 4.2 Add "Always Allow" button (ActionsRow 2, SecondaryButton) and `always:` interaction handling to Discord provider
- [x] 4.3 Add "Always Allow" button and `always:` action handling to Slack provider
- [x] 4.4 Update all channel `approvalPending.ch` types from `chan bool` to `chan approval.ApprovalResponse`

## 5. Integration

- [x] 5.1 Update `wrapWithApproval` signature to accept `*approval.GrantStore` and add grant check/record logic in `internal/app/tools.go`
- [x] 5.2 Add `GrantStore *approval.GrantStore` field to `App` struct in `internal/app/types.go`
- [x] 5.3 Wire `GrantStore` creation and pass to `wrapWithApproval` in `internal/app/app.go`

## 6. Test Updates

- [x] 6.1 Update `internal/approval/approval_test.go` — mock provider and assertions for `ApprovalResponse`
- [x] 6.2 Update `internal/approval/headless_test.go` — assertions for `ApprovalResponse`
- [x] 6.3 Update `internal/approval/tty_test.go` — assertions for `ApprovalResponse`
- [x] 6.4 Update `internal/channels/telegram/approval_test.go` — add AlwaysAllow test case
- [x] 6.5 Update `internal/channels/discord/approval_test.go` — assertions for `ApprovalResponse`
- [x] 6.6 Update `internal/channels/slack/approval_test.go` — assertions for `ApprovalResponse`
- [x] 6.7 Update `internal/gateway/server_test.go` — channel type and import for `ApprovalResponse`
