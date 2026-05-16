## MODIFIED Requirements

### Requirement: Sessions page with session list
The cockpit SHALL include a Sessions page showing session key and relative last update time, ordered newest-first by `UpdatedAt`.

#### Scenario: Rendered sessions-page text stays plain and single-line
- **WHEN** session keys or configured-source error text contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Sessions page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
