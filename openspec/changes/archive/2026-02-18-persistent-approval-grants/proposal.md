## Why

The current approval system requires user confirmation for every single dangerous tool invocation, creating friction when the same tool is invoked repeatedly within a session. Users need an "Always Allow" option that grants persistent per-session, per-tool approval so subsequent calls to the same tool are auto-approved without re-prompting.

## What Changes

- Add `ApprovalResponse` struct to replace bare `bool` returns from `Provider.RequestApproval`, carrying both `Approved` and `AlwaysAllow` fields
- Add in-memory `GrantStore` that tracks session+tool grants (cleared on restart)
- Update all approval providers (Composite, Headless, TTY, Gateway) to return `ApprovalResponse`
- Add "Always Allow" button/option to all channel UIs (Telegram, Discord, Slack)
- Update Gateway WebSocket protocol to support `alwaysAllow` field in `approval.response`
- Modify `wrapWithApproval` to check `GrantStore` before prompting and record grants on `AlwaysAllow` responses
- Wire `GrantStore` into the `App` struct and tool approval wrapping

## Capabilities

### New Capabilities
- `persistent-approval-grant`: In-memory per-session, per-tool grant store with Grant/IsGranted/Revoke/RevokeSession operations

### Modified Capabilities
- `approval-policy`: Provider interface returns `ApprovalResponse` instead of `bool`; `wrapWithApproval` checks GrantStore before prompting
- `channel-approval`: All channels add a third "Always Allow" button/option; response channels carry `ApprovalResponse` instead of `bool`

## Impact

- **Core**: `internal/approval/` — new type `ApprovalResponse`, interface change on `Provider` and `GatewayApprover`
- **Channels**: `internal/channels/{telegram,discord,slack}/approval.go` — UI button additions and response type changes
- **Gateway**: `internal/gateway/server.go` — `pendingApprovals` type change, `handleApprovalResponse` JSON field addition
- **App**: `internal/app/{tools,types,app}.go` — `GrantStore` wiring, `wrapWithApproval` signature change
- **Tests**: All approval-related test files updated for `ApprovalResponse` type
- **Backward compatible**: Companion WebSocket clients that omit `alwaysAllow` default to `false`
