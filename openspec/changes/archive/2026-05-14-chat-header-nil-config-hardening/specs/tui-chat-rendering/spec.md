## MODIFIED Requirements

### Requirement: Parts-based layout agreement
The `View()` method and `recalcLayout()` method SHALL use the same parts structure so that measured heights always match rendered output. The viewport height SHALL be computed by subtracting the measured heights of all fixed parts (header, turn status strip, composer or approval card, help footer) and separators from the terminal height.

#### Scenario: Header rendering fails closed when config is unavailable
- **WHEN** the chat header renders with a nil config pointer
- **THEN** it SHALL not panic
- **AND** it SHALL fall back to the existing default provider/model labels
