## MODIFIED Requirements

### Requirement: Parts-based layout agreement
The `View()` method and `recalcLayout()` method SHALL use the same parts structure so that measured heights always match rendered output. The viewport height SHALL be computed by subtracting the measured heights of all fixed parts (header, turn status strip, composer or approval card, help footer) and separators from the terminal height.

#### Scenario: Header and turn strip stay single-line on narrow terminals
- **WHEN** the chat header or turn status strip renders on a narrow terminal
- **THEN** each bar SHALL remain a single rendered line
- **AND** each bar SHALL clamp to the available terminal width instead of wrapping
