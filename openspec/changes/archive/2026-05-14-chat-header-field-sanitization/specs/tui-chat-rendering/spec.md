## MODIFIED Requirements

### Requirement: Parts-based layout agreement
The `View()` method and `recalcLayout()` method SHALL use the same parts structure so that measured heights always match rendered output. The viewport height SHALL be computed by subtracting the measured heights of all fixed parts (header, turn status strip, composer or approval card, help footer) and separators from the terminal height.

#### Scenario: Header display fields stay plain and single-line
- **WHEN** provider, model, or session-key display text contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the chat header SHALL strip those control sequences
- **AND** it SHALL normalize the display text to a single line before rendering
