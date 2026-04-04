## ADDED Requirements

### Requirement: Extended color palette with surface tokens
The theme SHALL define surface color tokens: `Surface0` (deepest background), `Surface1` (card background), `Surface2` (elevated surface), `Surface3` (highest surface). These SHALL complement the existing palette in `internal/cli/tui/styles.go`.

#### Scenario: Surface colors defined
- **WHEN** theme package is imported
- **THEN** `Surface0`, `Surface1`, `Surface2`, `Surface3` SHALL be available as `lipgloss.Color` values

### Requirement: Text and border color tokens
The theme SHALL define text tokens (`TextPrimary`, `TextSecondary`, `TextTertiary`) and border tokens (`BorderFocused`, `BorderDefault`, `BorderSubtle`).

#### Scenario: Text and border colors defined
- **WHEN** theme package is imported
- **THEN** all text and border color constants SHALL be available as `lipgloss.Color` values

### Requirement: Unicode icon constants
The theme SHALL define unicode icon constants for sidebar navigation: Chat (`◉`), Settings (`⚙`), Tools (`⚡`), Status (`◍`), Sessions (`◈`).

#### Scenario: Icons available
- **WHEN** icons are referenced from sidebar
- **THEN** each icon SHALL be a single-character string constant

### Requirement: Enhanced logo renderer
The theme SHALL provide a logo rendering function that produces the squirrel mascot ASCII art with color gradient (body in Primary purple, eyes in Foreground white).

#### Scenario: Logo rendering
- **WHEN** `RenderLogo()` is called
- **THEN** output SHALL contain the squirrel ASCII art with lipgloss color styling

### Requirement: No import from cockpit packages
The theme package SHALL only import from the Go standard library and `lipgloss`. It SHALL NOT import from other cockpit subpackages to prevent import cycles.

#### Scenario: Import safety
- **WHEN** theme package is compiled
- **THEN** imports SHALL be limited to stdlib and `github.com/charmbracelet/lipgloss`

### Requirement: Semantic color aliases
The theme package SHALL provide semantic color aliases that map to existing brand/status colors: `Danger` (alias for Error red), `Info` (alias for Highlight blue), `Selection` (alias for Accent green).

#### Scenario: Danger alias resolves to error color
- **WHEN** code references `theme.Danger`
- **THEN** it resolves to the same hex value as `theme.Error` (#EF4444)

#### Scenario: Info alias resolves to highlight color
- **WHEN** code references `theme.Info`
- **THEN** it resolves to the same hex value as `theme.Highlight` (#3B82F6)

### Requirement: Reduced border-heavy styles
The shared styles package SHALL provide spacing/badge/pill-based style helpers alongside existing border styles. New transcript item renderers SHALL prefer spacing over borders for visual hierarchy.

#### Scenario: Badge style available
- **WHEN** a renderer needs a status badge (e.g., tool state icon)
- **THEN** a `BadgeStyle(color)` helper returns a pill-shaped lipgloss style with padding and no border

#### Scenario: Divider style available
- **WHEN** a renderer needs a section separator
- **THEN** a `DividerStyle` renders a subtle horizontal line using spacing rather than a box border
