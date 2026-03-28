## Why

The current `lango` TUI is a single-column chat interface (header/strip/viewport/footer). Users want a richer experience with side navigation, integrated settings, and tool browsing — similar to charmbracelet/crush. To introduce this safely without breaking the existing stable TUI, we add an experimental `lango cockpit` subcommand that wraps the unmodified ChatModel in a 2-panel layout (sidebar + chat).

## What Changes

- Add `lango cockpit` experimental subcommand as a separate entry point
- Implement a 2-panel layout: left sidebar (navigation display) + right panel (existing ChatModel)
- ChatModel remains completely unmodified — cockpit wraps it via `childModel` interface with consume-or-forward message delegation
- Add cockpit-specific theme extension (Surface0-3 color tokens, unicode icons, enhanced logo)
- Add non-interactive sidebar component (display-only in this change; interactive navigation deferred to future changes)
- Wire runtime: `SetProgram` delegation, approval fallback type assertion, synthetic `WindowSizeMsg` on sidebar toggle

## Capabilities

### New Capabilities
- `cockpit-shell`: Root `cockpit.Model` tea.Model with 2-panel layout (sidebar + child), consume-or-forward message delegation, `childModel` interface, `Ctrl+B` sidebar toggle with synthetic resize re-propagation, `SetProgram` delegation, responsive sidebar width
- `cockpit-sidebar`: Non-interactive sidebar component displaying menu items with unicode icons and active highlight, collapsible (20ch full / 3ch icon-only)
- `cockpit-theme`: Extended color palette (Surface0-3, TextPrimary/Secondary/Tertiary, BorderFocused/Default/Subtle), unicode icon constants, enhanced squirrel ASCII logo with gradient coloring

### Modified Capabilities
<!-- No existing spec requirements are changing. The existing tui-cockpit-layout spec
     continues to govern the `lango` default entry point. cockpit is a separate command. -->

## Impact

- **New package**: `internal/cli/cockpit/` with subpackages `sidebar/` and `theme/`
- **Modified file**: `cmd/lango/main.go` — adds `cockpitCmd` Cobra subcommand and `runCockpit()` function
- **Dependencies**: No new external dependencies; uses existing bubbletea, lipgloss, bubbles
- **Existing code**: Zero modifications to `internal/cli/chat/`, `internal/cli/tui/`, or `internal/cli/settings/`
- **Risk**: ChatModel header/help bar render in reduced width; mitigated by existing `truncateSessionKey()` and width clamping (`max(w-N, 10)`)
