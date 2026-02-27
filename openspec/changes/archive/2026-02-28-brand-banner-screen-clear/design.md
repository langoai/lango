## Context

The Lango CLI lacks visual brand identity. TUI commands (`lango settings`, `lango onboard`) launch without clearing previous terminal output, and there is no mascot or version information displayed. The `lango serve` command also starts without any visual banner.

The `internal/cli/tui` package already provides shared styles (colors, typography) used across all TUI screens. The banner component extends this package with a reusable brand element.

## Goals / Non-Goals

**Goals:**
- Provide a reusable banner component in `internal/cli/tui` with squirrel mascot art
- Inject version/build/profile info via setter pattern (avoiding import cycles with `cmd/lango/main.go`)
- Clear screen on TUI launch for clean presentation
- Display serve banner before server startup log output

**Non-Goals:**
- Animated or dynamic banner content
- Configurable mascot art or theming
- Banner display on non-TUI commands (e.g., `lango config list`)

## Decisions

**1. Setter pattern for version injection**
Version/BuildTime live only in `cmd/lango/main.go`. Rather than passing them through constructors or using build-tag globals, package-level setters (`SetVersionInfo`, `SetProfile`) keep the API simple and avoid import cycles. This is the same pattern used by logging packages.

**2. `tea.ClearScreen` in Init()**
Bubbletea's `tea.ClearScreen` command is the idiomatic way to clear the terminal before rendering. Added to both `Editor.Init()` and `Wizard.Init()`.

**3. `lipgloss.JoinHorizontal` for art layout**
The squirrel art and info text are joined side-by-side using lipgloss horizontal join, which handles ANSI-aware width calculation correctly.

**4. Three banner variants**
- `Banner()` — raw banner (onboard wizard title)
- `BannerBox()` — banner wrapped in rounded border (settings welcome)
- `ServeBanner()` — banner with separator line (serve command stdout)

## Risks / Trade-offs

- [Wide Unicode characters] → The squirrel art uses block characters that may render differently across terminal emulators. Mitigated by using widely-supported Unicode block elements.
- [Global state for version] → Package-level vars are set once at startup and read-only thereafter. No concurrency risk in practice.
