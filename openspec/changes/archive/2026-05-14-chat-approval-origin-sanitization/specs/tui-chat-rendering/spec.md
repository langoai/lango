## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval surfaces sanitize displayed channel origin text
- **WHEN** a chat approval surface renders channel-origin text extracted from a session key that contains ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed origin text to a single line before rendering it
