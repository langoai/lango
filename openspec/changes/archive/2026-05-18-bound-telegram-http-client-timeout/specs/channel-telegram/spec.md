## MODIFIED Requirements

### Requirement: Telegram bot connection

The system SHALL connect to Telegram using the Bot API with a provided bot token. When no custom HTTP client is provided, Bot API requests SHALL use a default HTTP client with a finite timeout greater than the channel's long-poll update timeout.

#### Scenario: Default HTTP client has bounded timeout

- **WHEN** the Telegram channel is created without `Config.HTTPClient`
- **THEN** the Bot API client SHALL use an HTTP client with a finite timeout
- **AND** the timeout SHALL be greater than the 60 second long-poll update timeout

#### Scenario: Custom HTTP client is preserved

- **WHEN** the Telegram channel is created with `Config.HTTPClient`
- **THEN** the provided HTTP client SHALL be used unchanged
