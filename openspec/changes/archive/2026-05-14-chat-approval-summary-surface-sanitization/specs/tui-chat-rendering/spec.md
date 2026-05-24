## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval surfaces keep summaries plain and single-line
- **WHEN** any chat approval surface renders a summary containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the summary to a single line before displaying it
