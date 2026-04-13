## Context

The TUI chat (`internal/cli/chat/`) uses bubbletea with a viewport, textarea input, status bar, and help bar. Three rendering bugs exist:

1. **Input box border triplication**: `textarea.SetWidth(terminalWidth)` causes the border to wrap, producing 3 visible borders.
2. **AI response not displayed**: `DoneMsg` handler discards stream buffer on non-success outcomes and ignores `ResponseText` for non-streaming models.
3. **Excessive blank lines**: `render()` prepends `\n` to each entry, accumulating leading blank space.

Additionally, `View()` and `recalcLayout()` use independent height assumptions (hardcoded constants vs. actual rendered heights), causing layout drift.

## Goals / Non-Goals

**Goals:**
- Eliminate all three rendering bugs with verifiable unit tests.
- Unify View() and recalcLayout() so they share the same parts structure.
- Preserve assistant partial responses on error/cancel/timeout.
- Support resize reflow for assistant markdown content.

**Non-Goals:**
- Full chat architecture refactor (slash commands, TurnRunner integration).
- Performance optimization (markdown caching on width change).
- Runtime package changes (internal/app/, internal/graph/).

## Decisions

### 1. Parts-based View/Layout model
**Decision**: Both `View()` and `recalcLayout()` build the same `[]string` parts list. `View()` joins them; `recalcLayout()` measures their `lipgloss.Height()` sum.
**Rationale**: Eliminates the root cause of layout disagreement — hardcoded height constants that drift from actual rendered output. Alternatives: (a) keep constants and adjust — fragile, recurs on any component change; (b) measure in View() only — can't size viewport before rendering.

### 2. Block-join rendering
**Decision**: `render()` collects entries into `[]string` blocks and joins with `"\n\n"` instead of prepending `"\n"` to each entry.
**Rationale**: `strings.Join` produces zero leading blank lines and consistent inter-block spacing. The old approach accumulated `\n` prefixes, growing blank space proportional to entry count.

### 3. Single `appendAssistant()` entry point
**Decision**: All assistant messages (stream finalization, non-streaming ResponseText, error fallback) go through `appendAssistant(raw)` which stores both rendered content and raw markdown.
**Rationale**: Single path ensures rawContent is always populated for resize reflow. Eliminates the previous split between `finalizeStream()` and `finalizeWithText()` where only one stored raw content.

### 4. Input width -2 margin
**Decision**: `SetWidth(width)` internally sets `textarea.SetWidth(max(width-2, 10))`.
**Rationale**: The textarea's `RoundedBorder` adds 2 characters (left+right). Without the margin, the border wraps to the next line, producing the visible triplication. The minimum clamp (10) prevents panics at extremely narrow terminals.

### 5. DoneMsg 3-rule processing
**Decision**: Process in strict order: (1) finalize stream if non-empty, (2) else use ResponseText for non-streaming, (3) add error system message with deduplication.
**Rationale**: Rules are independent and composable. Rule 1 preserves partial responses even on failure. Rule 2 supports non-streaming model backends. Rule 3 adds error context without duplicating the response text already shown. Deduplication compares raw content to avoid styled-text comparison issues.

## Risks / Trade-offs

- **[Performance]** `render()` re-renders all assistant markdown on every call via `renderMarkdown(rawContent, contentWidth())`. → Acceptable for correctness-first approach; caching (re-render only on width change) is a future optimization if profiling shows need.
- **[Glamour overhead]** Each `renderMarkdown()` call creates a new `glamour.TermRenderer`. → Same as existing behavior; pooling is out of scope.
- **[Border assumption]** The -2 margin assumes `RoundedBorder` adds exactly 2 chars. → This is stable in lipgloss; if border style changes, the margin constant would need updating.
