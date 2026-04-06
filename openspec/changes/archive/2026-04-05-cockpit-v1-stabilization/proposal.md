## Why

Cockpit v1 overhaul shipped with rendering overflow bugs on narrow terminals and 0% test coverage on 6 new source files (chat package at 41.8%). Before adding new features (Phase 2+), these quality gaps must be closed to prevent regressions from compounding.

## What Changes

- Fix help bar rendering overflow on narrow terminals (`statusbar.go` — ignores width parameter)
- Fix approval strip content overflow on narrow terminals (`approval_strip.go` — byte-slice truncation breaks Unicode)
- Fix tasks page hardcoded column widths misaligning on narrow cockpit (`pages/tasks.go` — `"%-10s %-30s %-10s %s"`)
- Standardize all display string truncation to use `ansi.Truncate` (ANSI/Unicode-safe, no byte slicing)
- Extract `msgSender` interface in `bridge.go` for testability (unexported, `*tea.Program` still satisfies it)
- Add comprehensive tests for all 6 untested v1 files: approval_dialog, approval_strip, bridge, render_tool, render_thinking, taskstrip
- Add tests for chatview.go new transcript methods (appendToolStart, finalizeToolResult, appendThinking, finalizeThinking)
- Add tests for statusbar.go rendering functions and tasks page lifecycle

## Capabilities

### New Capabilities

(none — this is a stabilization/hardening change, no new user-facing behavior)

### Modified Capabilities

(none — bug fixes and test additions do not change spec-level requirements; they bring the implementation into compliance with existing specs)

## Impact

- `internal/cli/chat/`: 7 files modified, 8 new test files created (~95 new test cases)
- `internal/cli/cockpit/pages/`: 1 file modified, 1 new test file created
- `github.com/charmbracelet/x/ansi` promoted from indirect to direct dependency (already in go.sum)
- `internal/cli/chat/bridge.go`: function signature change (`*tea.Program` → `msgSender` interface) — backward compatible, call sites unchanged
