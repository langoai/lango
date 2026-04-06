## Context

Cockpit v1 shipped with 6 new source files at 0% test coverage and 3 rendering bugs on narrow terminals. The chat package was at 41.8% coverage. This stabilization pass fixes the bugs and adds comprehensive tests before Phase 2 (Channel-Aware Cockpit) adds more complexity.

## Goals / Non-Goals

**Goals:**
- Fix 3 confirmed rendering overflow bugs (help bar, approval strip, tasks page)
- Achieve test coverage for all v1 new files
- Establish truncation and rendering safety patterns for future TUI work

**Non-Goals:**
- New user-facing features
- Architecture changes or module restructuring
- Performance optimization (Phase 7 scope)

## Decisions

### D1. ansi.Truncate for all display string truncation

All visual string truncation uses `ansi.Truncate(s, width, tail)` from `github.com/charmbracelet/x/ansi` (already a transitive dependency via lipgloss).

**Alternative considered:** Custom `truncateToWidth` helper using rune iteration with `lipgloss.Width()`. Rejected because `ansi.Truncate` is battle-tested, ANSI-escape-safe, and already available.

**Alternative considered:** `lipgloss.MaxWidth()` style property. Rejected because it wraps text to the next line rather than truncating, which breaks single-line surfaces like help bar and approval strip.

### D2. Plain → truncate → style ordering

For single-line surfaces (help bar, approval strip, task strip), truncation happens on plain text before style application. This avoids cutting ANSI escape sequences mid-sequence.

For already-styled strings (e.g., HelpBar output which contains KeyBadge ANSI codes), `ansi.Truncate` is safe because it preserves escape sequences.

### D3. msgSender interface for bridge testability

`bridge.go` extracted an unexported `msgSender` interface (`Send(msg tea.Msg)`) and changed `enrichRequest` from `*tea.Program` to `msgSender`. This allows test injection via `mockSender` without requiring a running Bubble Tea program.

**Alternative considered:** Testing via ChatModel.Update integration. Rejected because it would overlap with chatview tests and couple bridge tests to the full model lifecycle.

`*tea.Program` already satisfies `msgSender`, so all call sites remain unchanged.

### D4. Tasks page narrow/wide column format split

Tasks page uses two distinct row formats based on `width >= 50` threshold:
- **Wide** (>= 50): ID + Prompt + Status + Elapsed — 4 columns with prompt taking remaining width
- **Narrow** (< 50): ID + Prompt + Status — 3 columns, elapsed hidden, more prompt space

Fixed-element widths are subtracted first, then remaining width goes to prompt. This prevents proportional calculations from producing overflows.

### D5. Test-only state reset via t.Cleanup

Package-level dialog state (`dialogScrollOffset`, `dialogSplitMode`) is reset in test files using `t.Cleanup()`, not via production `resetDialogState()` functions. Test-only concerns stay in `_test.go`.

## Risks / Trade-offs

**[R1] `ansi.Truncate` becomes direct dependency** — Promoted from indirect to direct in go.mod. Low risk: it's maintained by the Charm team and already pulled in via lipgloss. → Mitigation: Pin to same version lipgloss uses.

**[R2] Tasks page width=50 threshold is arbitrary** — Could be too narrow or too wide for some layouts. → Mitigation: The constant is defined once (`taskNarrowThreshold = 50`), easy to adjust.

**[R3] Parallel agents on same branch without worktree isolation** — Multiple agents modified different files concurrently. → Mitigation: Write surface analysis verified zero file overlap between all 7 units. All agents reported clean builds.
