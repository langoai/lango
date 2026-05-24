## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Inline approval strip keeps summary plain and single-line
- **WHEN** the Tier 1 inline approval strip renders a summary containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the summary to a single line before truncating it into the strip
