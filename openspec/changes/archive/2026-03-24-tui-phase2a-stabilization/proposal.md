## Why

The TUI chat interface has three blocking UX bugs that make it unusable for extended sessions: (1) ANSI CPR responses leak into the textarea as garbage characters like `[43;84R`, (2) mouse wheel scrolling doesn't work despite viewport support, and (3) WARN-level log output corrupts the alt-screen display. All three share the same root files and must be fixed together.

## What Changes

- Add `tea.WithMouseCellMotion()` to bubbletea program options to enable mouse wheel event delivery to the viewport
- Redirect TUI-mode logging from `os.Stderr` to a file at `<DataRoot>/chat.log` to prevent alt-screen corruption
- Add a CPR (Cursor Position Report) filter state machine in `ChatModel.Update()` that intercepts and discards `ESC[row;colR` sequences before they reach the textarea, with a 50ms timeout to avoid blocking real Escape key presses

## Capabilities

### New Capabilities
- `tui-cpr-filter`: CPR sequence detection and filtering state machine for terminal input sanitization

### Modified Capabilities
- `tui-chat-rendering`: Mouse scroll enablement and log output path change in TUI program initialization

## Impact

- `cmd/lango/main.go`: Program options and logging config changes in `runChat()`
- `internal/cli/chat/chat.go`: New CPR filter types, fields, and methods on `ChatModel`
- `internal/cli/chat/chat_test.go`: New test cases for CPR filter behavior
- No dependency changes, no API changes, no breaking changes
