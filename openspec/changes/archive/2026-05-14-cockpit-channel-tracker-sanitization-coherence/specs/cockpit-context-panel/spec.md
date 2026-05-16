## MODIFIED Requirements

### Requirement: Channel status section in context panel
The context panel SHALL display a "Channels" section showing each channel's connection status (connected/disconnected indicator), name, and message count.

#### Scenario: Channel snapshot labels are replay-safe
- **WHEN** channel names contain ANSI/OSC escape sequences or embedded newlines before entering the channel snapshot
- **THEN** the channel tracker SHALL strip those control sequences
- **AND** it SHALL normalize the stored channel snapshot text to a single line before replay
