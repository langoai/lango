## ADDED Requirements

### Requirement: Double-close safety for private startTyping
The private `startTyping` functions in Discord and Telegram channel adapters SHALL use `sync.Once` to ensure the returned stop function is safe to call multiple times without panicking.

#### Scenario: Stop function called once
- **WHEN** the stop function returned by `startTyping` is called once
- **THEN** the typing indicator goroutine SHALL be stopped

#### Scenario: Stop function called multiple times
- **WHEN** the stop function returned by `startTyping` is called more than once
- **THEN** it SHALL NOT panic
- **AND** subsequent calls SHALL be no-ops
