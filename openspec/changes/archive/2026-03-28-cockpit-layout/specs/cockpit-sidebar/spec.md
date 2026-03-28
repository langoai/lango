## ADDED Requirements

### Requirement: Sidebar displays menu items with active highlight
The sidebar SHALL render a vertical list of menu items, each with a unicode icon and label. The currently active item SHALL be visually distinguished with accent color and a left border indicator.

#### Scenario: Render with Chat active
- **WHEN** sidebar renders with active page "Chat"
- **THEN** the Chat item SHALL display with Primary color icon, Bold label, and left accent bar
- **AND** all other items SHALL display with Muted color

### Requirement: Sidebar is non-interactive in Change-1
The sidebar SHALL NOT consume any key events. All key events — including Up, Down, Enter — SHALL pass through to the parent cockpit model (which forwards them to the child).

#### Scenario: Key events pass through
- **WHEN** sidebar receives any KeyMsg
- **THEN** sidebar SHALL return the message unhandled (no state change, no command)

### Requirement: Sidebar supports visibility toggle
The sidebar SHALL support `SetVisible(bool)` to show or hide. When hidden, `View()` SHALL return an empty string.

#### Scenario: Sidebar hidden
- **WHEN** sidebar is not visible
- **THEN** `View()` SHALL return `""`

### Requirement: Sidebar supports height adjustment
The sidebar SHALL support `SetHeight(int)` to match the terminal height. The sidebar panel SHALL fill the full terminal height.

#### Scenario: Height matches terminal
- **WHEN** `SetHeight(40)` is called
- **THEN** sidebar View SHALL render with height 40

### Requirement: Sidebar fixed width
The sidebar SHALL have a fixed width of 20 characters when fully displayed.

#### Scenario: Full width
- **WHEN** sidebar is visible
- **THEN** the rendered width SHALL be exactly 20 characters
