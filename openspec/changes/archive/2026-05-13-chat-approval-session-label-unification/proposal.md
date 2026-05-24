## Why

The chat approval surfaces already use `allow session` wording almost everywhere, but the fullscreen diff approval dialog still abbreviates the `s` action as just `session`. That leaves the same action labeled differently across approval surfaces.

## What Changes

- Change the fullscreen diff approval dialog help label from `session` to `allow session`.
- Add regression coverage for the diff-dialog action bar.
- Update the chat-rendering spec to require the unified session-approval wording.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Approval dialog action-bar wording becomes consistent for the session-approval action.

## Impact

- Affected code: `internal/cli/chat/approval_dialog.go`, `internal/cli/chat/approval_dialog_test.go`
- Affected specs: `openspec/specs/tui-chat-rendering/spec.md`
