## 1. Bug Fixes

- [x] 1.1 Fix help bar width overflow — `statusbar.go` renderHelpBar to use `ansi.Truncate(bar, w, "")` instead of ignoring width
- [x] 1.2 Fix approval strip overflow — `approval_strip.go` compute available summary space, truncate plain text with `ansi.Truncate` before styling
- [x] 1.3 Fix tasks page column widths — `pages/tasks.go` replace hardcoded `"%-10s %-30s %-10s %s"` with dynamic wide/narrow format split at width=50

## 2. Testability Improvements

- [x] 2.1 Extract `msgSender` interface in `bridge.go` — unexported, `*tea.Program` satisfies it, call sites unchanged

## 3. Test Coverage — Renderers

- [x] 3.1 Add `render_tool_test.go` — toolStateVisual all states, renderToolBlock edge cases (truncation, empty, narrow, zero width, Unicode)
- [x] 3.2 Add `render_thinking_test.go` — renderThinkingBlock active/done/empty/unknown, renderPendingIndicator
- [x] 3.3 Add `approval_strip_test.go` — normal/narrow/very narrow/zero width, long summary, empty summary, Korean chars, single-line height guarantee
- [x] 3.4 Add `approval_dialog_test.go` — normal/narrow/short/minimal size, diff/scroll/split, params, key handling, risk color
- [x] 3.5 Add `statusbar_test.go` — renderHelpBar all states + narrow width regression, turnStateCopy, renderHeader, renderTurnStrip

## 4. Test Coverage — Bridge and Task Strip

- [x] 4.1 Add `bridge_test.go` — nil sender, callback wiring, OnChunk preservation, ToolStartedMsg delivery, thinking boundary + duration
- [x] 4.2 Add `taskstrip_test.go` — nil manager, empty snapshots, single running, long prompt, narrow/zero width, completed elapsed, sort order

## 5. Test Coverage — Transcript Operations

- [x] 5.1 Extend `chatview_test.go` — appendToolStart, finalizeToolResult (success/error/output/no-match), appendThinking, finalizeThinking (done/summary/no-active), clear, appendSystem

## 6. Test Coverage — Tasks Page

- [x] 6.1 Add `pages/tasks_test.go` — title, nil lister, empty tasks, cursor nav/clamp/highlight, ID/prompt truncation, activate/deactivate, narrow/wide width regression

## 7. Verification

- [x] 7.1 `go build ./...` passes
- [x] 7.2 `go test ./...` passes
- [x] 7.3 `go vet ./...` passes
