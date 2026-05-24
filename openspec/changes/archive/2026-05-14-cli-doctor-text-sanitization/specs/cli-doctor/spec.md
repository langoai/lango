## MODIFIED Requirements

### Requirement: Doctor Command Entry Point
The system SHALL include RunLedger diagnostics in the `lango doctor` command output and help text.

#### Scenario: Doctor output text stays plain and single-line
- **WHEN** doctor check names, messages, details, fix actions, or structured trace metadata contain ANSI/OSC escape sequences or embedded newlines before rendering
- **THEN** the doctor command SHALL strip those control sequences
- **AND** it SHALL normalize both TUI and JSON output text to a single line before display or serialization
