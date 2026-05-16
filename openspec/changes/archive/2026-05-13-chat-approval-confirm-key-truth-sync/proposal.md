## Why

Critical-risk approval confirmation already supports both `a` and `s` as double-press actions, but the visible confirm prompt is hard-coded to `Press 'a' again...` in both the inline strip and fullscreen dialog. That means the UI can tell the operator to press the wrong key when they chose the session-grant path.

## What Changes

- Render the actual pending confirm key (`a` or `s`) in both approval confirm prompts.
- Add regressions for the inline strip and fullscreen dialog session-confirm path.
- Update the chat-rendering spec to require key-accurate confirm prompts.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Approval confirm prompts must name the actual pending action key.

## Impact

- Affected code: `internal/cli/chat/approval.go`, `internal/cli/chat/approval_strip.go`, `internal/cli/chat/approval_dialog.go`
- Affected tests/specs: `internal/cli/chat/approval_strip_test.go`, `internal/cli/chat/approval_dialog_test.go`, `openspec/specs/tui-chat-rendering/spec.md`
