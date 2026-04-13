## Context

The TUI chat interface (`cmd/lango/main.go` + `internal/cli/chat/`) has three blocking UX bugs that share the same root files:

1. **CPR leak**: Some terminals emit `ESC[row;colR` (Cursor Position Report) when alt-screen or mouse mode activates. Bubbletea's input reader can split this across multiple reads, causing the escape to be parsed as `KeyEscape` and the remainder `[43;84R` as individual `KeyRunes` that get inserted into the textarea.
2. **No mouse scroll**: `tea.NewProgram` is created with `tea.WithAltScreen()` only. Without `tea.WithMouseCellMotion()`, bubbletea never subscribes to mouse events, so `tea.MouseMsg` is never delivered to the viewport.
3. **Log corruption**: Logging uses `Writer: os.Stderr`. Bubbletea alt-screen does not capture stderr, so async goroutine logs (e.g., turnrunner trace recorder) visually corrupt the TUI.

## Goals / Non-Goals

**Goals:**
- Eliminate CPR garbage in textarea input
- Enable mouse wheel scrolling in chat viewport
- Prevent log output from corrupting the alt-screen TUI

**Non-Goals:**
- TUI redesign (single-column cockpit layout) — separate Phase 2b
- In-TUI log viewer panel — future work
- Full terminal compatibility audit — only CPR leak is addressed

## Decisions

### D1: CPR filter as Update()-level state machine
**Choice**: Intercept `tea.KeyMsg` in `ChatModel.Update()` with a 4-state FSM (`cprIdle → cprGotEsc → cprGotBracket → cprInParams`) before delegation to `handleKey()` or input component.

**Rationale**: The CPR sequence arrives as individual `KeyMsg` events due to chunked terminal reads. Filtering at the `Update()` entry point is the earliest interception point, preventing contamination of both key handlers and the textarea. A state machine is the simplest correct approach for a stateful multi-character sequence.

**Alternative considered**: Patching bubbletea's input reader — rejected because it's a third-party dependency and the fix must be localized.

### D2: 50ms timeout for CPR detection window
**Choice**: When ESC is received, start a 50ms `tea.Tick`. If the CPR sequence doesn't complete, flush buffered keys as normal input.

**Rationale**: Real Esc key presses must not be delayed indefinitely. 50ms is well within terminal round-trip time for CPR but imperceptible to human Esc key usage. This matches similar timeout values used in other terminal multiplexers (tmux uses 50ms escape-time by default).

### D3: Mouse via WithMouseCellMotion (not WithMouseAllMotion)
**Choice**: `tea.WithMouseCellMotion()` — tracks click, release, and wheel events only.

**Rationale**: `WithMouseAllMotion` also tracks hover/motion which generates excessive events and can interfere with text selection in some terminal emulators. Cell motion is sufficient for viewport scrolling.

### D4: Log redirect to file (not discard)
**Choice**: Redirect TUI logging to `<DataRoot>/chat.log` instead of discarding or using a ring buffer.

**Rationale**: Logs are essential for debugging TUI issues. File redirect is the simplest fix that preserves debuggability. The `logging.Init` `OutputPath` branch already handles `O_APPEND|O_CREATE|O_WRONLY` — no new code needed in the logging package.

## Risks / Trade-offs

- **[Risk] CPR filter adds latency to Esc key** → Mitigated by 50ms timeout; imperceptible in practice. Flush path replays buffered keys through normal handlers.
- **[Risk] Log file grows unbounded** → Acceptable for tactical fix. Users can delete or rotate manually. Future: add log rotation or size cap.
- **[Risk] CPR state machine complexity** → Minimized to 4 states with clear transitions. All paths tested. Non-CPR sequences flush correctly.
