## Purpose

Brand banner component providing the Lango squirrel mascot, version info, and profile display across CLI/TUI surfaces (settings welcome, onboard wizard, serve startup).

## Requirements

### Requirement: Banner component provides squirrel mascot with version info
The `tui` package SHALL provide a `Banner()` function that returns a string containing the squirrel mascot ASCII art alongside version, tagline, and profile information arranged horizontally.

#### Scenario: Banner displays version and profile
- **WHEN** `SetVersionInfo("0.4.0", "2026-01-01")` and `SetProfile("default")` are called before `Banner()`
- **THEN** the output SHALL contain "Lango v0.4.0", "Fast AI Agent in Go", and "profile: default"

### Requirement: BannerBox wraps banner in rounded border
The `tui` package SHALL provide a `BannerBox()` function that wraps the banner in a rounded border box styled with the Primary color.

#### Scenario: BannerBox has border characters
- **WHEN** `BannerBox()` is called
- **THEN** the output SHALL contain rounded border characters (e.g., "╭", "│")

### Requirement: ServeBanner includes separator line
The `tui` package SHALL provide a `ServeBanner()` function that renders the banner followed by a horizontal separator line using the Separator color.

#### Scenario: ServeBanner contains separator
- **WHEN** `ServeBanner()` is called
- **THEN** the output SHALL contain horizontal line characters ("─")

### Requirement: TUI screens clear terminal on launch
The settings editor and onboard wizard SHALL return `tea.ClearScreen` from their `Init()` method to clear previous terminal output.

#### Scenario: Settings editor clears screen
- **WHEN** the settings editor initializes
- **THEN** `Init()` SHALL return `tea.ClearScreen`

#### Scenario: Onboard wizard clears screen
- **WHEN** the onboard wizard initializes
- **THEN** `Init()` SHALL return `tea.ClearScreen`

### Requirement: Serve command prints banner before startup
The `lango serve` command SHALL print the serve banner to stdout after logging initialization and before starting the application.

#### Scenario: Serve displays banner with profile
- **WHEN** `lango serve` is executed
- **THEN** the serve banner SHALL be printed with the active profile name

### Requirement: Version injection via setter pattern
The banner component SHALL use package-level setter functions (`SetVersionInfo`, `SetProfile`) to receive version, build time, and profile information, avoiding import cycles with `cmd/lango/main.go`.

#### Scenario: Version defaults before injection
- **WHEN** no setter is called
- **THEN** version SHALL default to "dev" and profile SHALL default to "default"
