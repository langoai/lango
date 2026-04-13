## ADDED Requirements

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
