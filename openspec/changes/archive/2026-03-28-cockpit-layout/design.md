## Context

The current `lango` TUI is a single-column chat interface governed by the `tui-cockpit-layout` spec. Users want a multi-panel experience with side navigation, integrated settings, and tool browsing. To introduce this without breaking the existing stable TUI, we add `lango cockpit` as an experimental subcommand.

The existing `ChatModel` (`internal/cli/chat/chat.go`) owns its full layout: header, turn strip, viewport, footer, and approval card. Its `recalcLayout()` and `View()` are tightly coupled to this 4-part structure. Any modification risks breaking the existing TUI.

## Goals / Non-Goals

**Goals:**
- Deliver `lango cockpit` subcommand with 2-panel layout (sidebar + existing ChatModel)
- Zero modifications to existing ChatModel, Settings Editor, or shared TUI packages
- Establish `childModel` interface and consume-or-forward message delegation pattern
- Validate that ChatModel renders correctly at reduced width

**Non-Goals:**
- Interactive sidebar navigation (deferred to Change-2 when multiple pages exist)
- 3-panel layout with context panel (deferred to Change-3)
- Mouse zone interaction or clipboard support (deferred to Change-3)
- ChatModel headless mode or `renderParts()` refactoring (deferred to Change-3)
- Settings, Tools, Status, or Sessions pages (deferred to Change-2)

## Decisions

### D1: Wrap ChatModel without modification via `childModel` interface

ChatModel is used as-is through a `childModel` interface:
```go
type childModel interface {
    tea.Model
    SetProgram(p *tea.Program)
}
```

**Rationale**: ChatModel's `View()` and `recalcLayout()` are tightly coupled to header/strip/footer height calculations. Adding a `headless` flag would require modifying `recalcLayout()`, which risks breaking existing tests and the stable `lango` entry point. The interface approach allows future substitution and enables mock-based testing.

**Alternative rejected**: `SetHeadless(bool)` flag — too invasive; `recalcLayout()` would need conditional height logic.

### D2: Consume-or-forward message delegation

Cockpit `Update()` only consumes messages it handles (`WindowSizeMsg` for width reduction, `KeyMsg` for `Ctrl+B`). All other messages — including `ChunkMsg`, `DoneMsg`, `ErrorMsg`, `WarningMsg`, `ApprovalRequestMsg`, `SystemMsg`, remaining `KeyMsg`, `MouseMsg`, `spinner.TickMsg` — are forwarded to the child via the default case.

**Rationale**: ChatModel processes 6+ custom message types via `program.Send()`. A whitelist approach would silently drop messages when new types are added. The default-forward pattern is future-proof.

### D3: `lango cockpit` subcommand (not flag, not default)

Cockpit is a separate Cobra subcommand, not `--cockpit` flag on the root command.

**Rationale**: A subcommand provides cleaner separation, its own `--help`, and can be removed without affecting the root command. A flag would couple cockpit lifecycle to the main chat command.

### D4: Non-interactive sidebar in Change-1

The sidebar displays menu items with active highlight but accepts no key input. All keys except `Ctrl+B` go to ChatModel.

**Rationale**: With only one page (Chat), sidebar navigation is meaningless. Adding up/down/enter creates Enter key conflicts with chat submit. Focus model deferred to Change-2 when multiple pages justify it.

### D5: Synthetic WindowSizeMsg on sidebar toggle

After `Ctrl+B` toggles sidebar visibility, cockpit sends a synthetic `WindowSizeMsg` to the child with the new effective width.

**Rationale**: ChatModel only recalculates layout on `WindowSizeMsg` or state transitions. Without synthetic resize, the child's width would be stale until the next terminal resize event.

### D6: SetProgram delegation (not ChatModel accessor)

Cockpit exposes `SetProgram(p)` which delegates to `m.child.SetProgram(p)`. ChatModel is never directly exposed.

**Rationale**: Exposing `ChatModel()` accessor would leak implementation details and couple callers to the concrete type. The delegation pattern keeps the child abstract.

## Risks / Trade-offs

| Risk | Impact | Mitigation |
|------|--------|------------|
| ChatModel header truncation at narrow width | Visual degradation | `truncateSessionKey()` already handles this. Minimum effective width: 60ch (80ch terminal - 20ch sidebar). |
| Import cycle: cockpit → chat → cockpit | Build failure | Chat never imports cockpit. Page switching uses message types, not direct calls. |
| Approval banner at reduced width | Visual degradation | `max(width-4, 10)` clamping already works. 60ch effective width is sufficient. |
| `childModel` type assertion panic | Runtime crash | ChatModel satisfies `childModel` interface — compile-time verified via `var _ childModel = (*chat.ChatModel)(nil)`. |
| `tea.Model.Update()` returns `tea.Model`, not `childModel` | Type mismatch | `updated.(childModel)` assertion after each `Update()` call. Safe because ChatModel always returns `*ChatModel`. |
