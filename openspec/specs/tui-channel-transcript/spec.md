## Purpose

Capability spec for tui-channel-transcript. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Channel message transcript item
The chat transcript SHALL display channel-origin messages as `itemChannel` entries with per-channel colored badge, sender name, and sanitized message text.

#### Scenario: Telegram message displayed in transcript
- **WHEN** a `ChannelMessageMsg` with Channel="telegram" and SenderName="alice" arrives
- **THEN** the transcript appends an `itemChannel` entry rendered with a blue badge, "@alice", and the message text

#### Scenario: Discord message displayed in transcript
- **WHEN** a `ChannelMessageMsg` with Channel="discord" arrives
- **THEN** the transcript appends an `itemChannel` entry rendered with an indigo badge

#### Scenario: Empty sender name
- **WHEN** a `ChannelMessageMsg` with SenderName="" arrives
- **THEN** the rendered block does NOT contain an "@" sender prefix

### Requirement: Channel text sanitization
The channel message renderer SHALL strip ANSI/OSC escape sequences from external channel text before rendering to prevent terminal control injection.

#### Scenario: ANSI escape stripped
- **WHEN** channel message text contains ANSI escape sequences (e.g., `\x1b[31mred\x1b[0m`)
- **THEN** `ansi.Strip()` removes them before rendering, and the output contains only plain text

#### Scenario: Multiline text collapsed
- **WHEN** channel message text contains newline characters
- **THEN** newlines are replaced with spaces for single-line display

### Requirement: Channel metadata preserved in transcript
The `appendChannel` method SHALL store `sessionKey` and platform-specific `metadata` in the transcript item's `meta` map for future extensibility (thread grouping, origin jump, per-channel badge).

#### Scenario: SessionKey preserved
- **WHEN** a channel message is appended with sessionKey="telegram:123:456"
- **THEN** the transcript item's meta["sessionKey"] equals "telegram:123:456"
