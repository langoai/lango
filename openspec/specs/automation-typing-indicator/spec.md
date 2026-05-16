# automation-typing-indicator Specification

## Purpose
Define typing-indicator behavior for automation deliveries across supported channel adapters.
## Requirements
### Requirement: Automation typing indicator dispatch
The `channelSender` SHALL implement a `StartTyping(ctx, channel) (func(), error)` method that dispatches typing indicator requests to the appropriate channel adapter based on the delivery target format (`channel:id`).

#### Scenario: Telegram typing dispatch
- **WHEN** `StartTyping` is called with target `telegram:<chatID>`
- **THEN** the system SHALL call `telegram.Channel.StartTyping(ctx, chatID)` and return its stop function

#### Scenario: Discord typing dispatch
- **WHEN** `StartTyping` is called with target `discord:<channelID>`
- **THEN** the system SHALL call `discord.Channel.StartTyping(ctx, channelID)` and return its stop function

#### Scenario: Slack typing dispatch
- **WHEN** `StartTyping` is called with target `slack:<channelID>`
- **THEN** the system SHALL call `slack.Channel.StartTyping(channelID)` and return its stop function

#### Scenario: Missing target ID returns no-op
- **WHEN** `StartTyping` is called with a target that has no ID (e.g., `discord` without `:channelID`)
- **THEN** the system SHALL return a no-op stop function and no error

#### Scenario: Unknown channel returns no-op
- **WHEN** `StartTyping` is called with an unavailable channel
- **THEN** the system SHALL return a no-op stop function and no error

### Requirement: Stop function is always non-nil
The `StartTyping` method SHALL always return a non-nil stop function, even on error. Callers MUST NOT need nil checks before calling stop.

#### Scenario: Error returns no-op stop
- **WHEN** `StartTyping` encounters a parse error (e.g., invalid chat ID)
- **THEN** the system SHALL return a no-op stop function alongside the error
