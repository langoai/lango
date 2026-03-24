## 1. OpenSpec And UI Model

- [x] 1.1 Update the new change artifacts to capture the single-column cockpit redesign and modified TUI requirements
- [x] 1.2 Refactor transcript item modeling in `internal/cli/chat/` to typed items with unified assistant append helpers

## 2. Layout And Rendering

- [x] 2.1 Convert `ChatModel.View()` and `recalcLayout()` to a shared parts-based cockpit layout
- [x] 2.2 Rebuild transcript rendering with block-joined item renderers and assistant resize reflow
- [x] 2.3 Keep the composer visible but muted during streaming, and replace it with an approval interrupt card during approval

## 3. State And Approval Flow

- [x] 3.1 Introduce explicit turn state handling for idle, streaming, approving, cancelling, and failed
- [x] 3.2 Route DoneMsg, ErrorMsg, cancel, warning, and approval outcomes through unified transcript/status append helpers
- [x] 3.3 Preserve and scope CPR filtering so idle composer input is protected without interfering with approval interactions

## 4. Verification And Downstream Updates

- [x] 4.1 Expand `internal/cli/chat/` tests for transcript typing, layout sizing, resize reflow, approval layout, and CPR handling
- [x] 4.2 Update README and CLI docs to describe the cockpit-style TUI behavior accurately
- [x] 4.3 Run `go build ./...`, `go test ./...`, and complete OpenSpec verify/sync/archive for the change
