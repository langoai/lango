## MODIFIED Requirements

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Approval origin/badge matching uses the sanitized session-key prefix
- **WHEN** an approval surface receives a known channel prefix wrapped in ANSI/OSC escape sequences inside the session key
- **THEN** it SHALL strip those control sequences before matching the prefix to Telegram, Discord, or Slack
