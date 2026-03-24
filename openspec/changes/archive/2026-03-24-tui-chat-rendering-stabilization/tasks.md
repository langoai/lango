## 1. Chat View Rendering

- [x] 1.1 Add `rawContent` field to `chatEntry` struct in `chatview.go`
- [x] 1.2 Implement `contentWidth()` method returning `max(width-2, 10)`
- [x] 1.3 Implement `appendAssistant(raw)` helper that stores rawContent and rendered content
- [x] 1.4 Replace `finalizeStream(width)` with no-arg `finalizeStream()` that delegates to `appendAssistant`
- [x] 1.5 Rewrite `render()` to use block-join (`strings.Join(blocks, "\n\n")`) with resize reflow from rawContent

## 2. Input Width Safety

- [x] 2.1 Update `SetWidth()` in `input.go` to subtract 2 with minimum clamp of 10

## 3. Layout Unification

- [x] 3.1 Rewrite `View()` in `chat.go` to use parts-based `[]string` joined by `"\n"`
- [x] 3.2 Rewrite `recalcLayout()` to measure the same fixed parts via `lipgloss.Height()` and compute viewport remainder

## 4. Message Processing Rules

- [x] 4.1 Implement DoneMsg 3-rule processing: finalize stream → ResponseText fallback → system error with dedup
- [x] 4.2 Implement ErrorMsg partial-first: preserve stream before adding error system message
- [x] 4.3 Update streaming cancel handler to preserve partial stream via `finalizeStream()` + system message

## 5. Approval Layout

- [x] 5.1 Call `recalcLayout()` in ApprovalRequestMsg handler
- [x] 5.2 Add width clamp `max(width-4, 10)` in `renderApprovalBanner()`

## 6. Tests

- [x] 6.1 Add `chatview_test.go`: appendAssistant, finalizeStream, block-join, resize reflow, contentWidth, streaming block
- [x] 6.2 Add `chat_test.go`: DoneMsg success/non-streaming/failure/dedup, ErrorMsg partial/no-stream, layout height, approval recalc, view parts
- [x] 6.3 Verify `go build ./...` and `go test ./...` pass
