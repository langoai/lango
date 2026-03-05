# thinking-indicator Specification

## Purpose

Cross-cutting typing/thinking indicator capability for all channel adapters. Provides immediate visual feedback to users while the agent processes messages, using each platform's native mechanism.
## Requirements
### Requirement: Typing indicator during agent processing
All channel adapters SHALL show a typing or thinking indicator immediately when a user message is received, and SHALL stop the indicator when the agent response is ready.

#### Scenario: Indicator starts before handler
- **WHEN** a user sends a message to any channel adapter
- **THEN** the adapter SHALL activate the platform-native thinking indicator before invoking the message handler

#### Scenario: Indicator stops after handler completes
- **WHEN** the message handler returns (success or error)
- **THEN** the adapter SHALL stop the thinking indicator

#### Scenario: Indicator failure does not block response
- **WHEN** the thinking indicator API call fails
- **THEN** the adapter SHALL log a warning and continue processing the message normally

### Requirement: Double-close safety for private startTyping
The private `startTyping` functions in Discord and Telegram channel adapters SHALL use `sync.Once` to ensure the returned stop function is safe to call multiple times without panicking.

#### Scenario: Stop function called once
- **WHEN** the stop function returned by `startTyping` is called once
- **THEN** the typing indicator goroutine SHALL be stopped

#### Scenario: Stop function called multiple times
- **WHEN** the stop function returned by `startTyping` is called more than once
- **THEN** it SHALL NOT panic
- **AND** subsequent calls SHALL be no-ops

