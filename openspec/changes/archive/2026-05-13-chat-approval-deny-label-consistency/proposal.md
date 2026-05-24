## Why

The chat approval surfaces currently use several deny-label variants such as `d`, `esc`, `d/esc`, and `[d]/Esc`. The behavior is the same, but the inconsistent casing and formatting make the approval UI feel less intentional than it should.

## What Changes

- Standardize chat approval deny labels on `d/Esc`.
- Add regression coverage for the visible approval surfaces that expose deny help.
- Update the chat-rendering spec to require consistent deny-key wording.

## Capabilities

### New Capabilities

### Modified Capabilities

- `tui-chat-rendering`: Approval deny affordances use one consistent `d/Esc` label across chat approval surfaces.

## Impact

- Affected code: `internal/cli/chat/approval.go`, `internal/cli/chat/approval_dialog.go`, `internal/cli/chat/statusbar.go`, `internal/cli/chat/approval_strip.go`
- Affected tests/specs: `internal/cli/chat/approval_dialog_test.go`, `internal/cli/chat/approval_strip_test.go`, `internal/cli/chat/statusbar_test.go`, `openspec/specs/tui-chat-rendering/spec.md`
