## Why

TUI commands (`lango settings`, `lango onboard`) display without clearing previous CLI output, making the interface look cluttered. Additionally, there is no brand identity on launch — no mascot or version info — resulting in a bland first impression. `lango serve` also lacks a startup banner.

## What Changes

- Add a squirrel mascot ASCII art banner with version info and profile name
- Clear screen on TUI launch (`settings`, `onboard`) for a clean start
- Display serve banner before server startup
- Replace the plain welcome box in settings with the branded banner box
- Replace the plain title in onboard wizard with the branded banner

## Capabilities

### New Capabilities
- `brand-banner`: Reusable banner component (squirrel mascot + version/profile info) with variants for TUI welcome, serve output, and boxed display

### Modified Capabilities

## Impact

- `internal/cli/tui/banner.go` — New banner component with setter pattern for version injection
- `cmd/lango/main.go` — Version info injection + serve banner output
- `internal/cli/settings/editor.go` — Screen clear on Init, banner box in welcome view
- `internal/cli/onboard/wizard.go` — Screen clear on Init, banner in title area
- `internal/cli/settings/settings.go` — Profile name injection before TUI launch
- `internal/cli/onboard/onboard.go` — Profile name injection before TUI launch
