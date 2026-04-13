## 1. OpenSpec And TUI Rendering Spec

- [x] 1.1 Capture the OSC 11 stability pass in proposal, design, and delta specs
- [x] 1.2 Update `tui-chat-rendering` and `tui-cpr-filter` requirements for explicit markdown style and OSC-safe input handling

## 2. Markdown And Input Guard

- [x] 2.1 Replace Glamour auto-style detection with a fixed dark style in TUI markdown rendering
- [x] 2.2 Extend the idle composer terminal-response guard to discard OSC responses while preserving non-matching sequences

## 3. Transcript Readability

- [x] 3.1 Refine transcript block renderers so user and assistant entries are visually distinct without heavy card UI
- [x] 3.2 Keep status and approval rows compact while preserving existing transcript flow and resize reflow behavior

## 4. Verification And Downstream Updates

- [x] 4.1 Expand `internal/cli/chat/` tests for OSC discard, CPR preservation, and transcript visual markers
- [x] 4.2 Update README and CLI docs where TUI behavior is described
- [x] 4.3 Run `go build ./...`, `go test ./...`, and complete OpenSpec verify/sync/archive for the change
