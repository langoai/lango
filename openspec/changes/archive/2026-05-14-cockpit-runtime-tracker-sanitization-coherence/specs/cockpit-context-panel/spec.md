## MODIFIED Requirements

### Requirement: Runtime status section in context panel
The context panel SHALL display a "Runtime" section showing the active agent, delegation count, and per-turn token usage when a turn is active. The section SHALL appear between Tool Stats and Channels.

#### Scenario: Runtime snapshot labels are replay-safe
- **WHEN** active-agent labels contain ANSI/OSC escape sequences or embedded newlines before entering the runtime snapshot
- **THEN** the runtime tracker SHALL strip those control sequences
- **AND** it SHALL normalize the stored runtime snapshot text to a single line before replay
