## ADDED Requirements

### Requirement: Slack progressive thinking indicator
The Slack channel SHALL post a "Thinking..." placeholder and periodically update it with elapsed time every 15 seconds in the format "_Thinking... (Xs)_".

#### Scenario: Placeholder posted on message receipt
- **WHEN** a user message is received in Slack
- **THEN** the channel SHALL post a "_Thinking..._" placeholder message

#### Scenario: Placeholder updated with elapsed time
- **WHEN** 15 seconds have elapsed since the placeholder was posted
- **THEN** the placeholder SHALL be updated to "_Thinking... (15s)_"

#### Scenario: Placeholder replaced with response
- **WHEN** the agent returns a successful response
- **THEN** the placeholder SHALL be edited to contain the formatted response

#### Scenario: Placeholder updated with error on failure
- **WHEN** the agent returns an error
- **THEN** the placeholder SHALL be edited to show the formatted error

### Requirement: Telegram progressive thinking indicator
The Telegram channel SHALL post a "Thinking..." placeholder message and periodically edit it with elapsed time. It SHALL fall back to typing indicators if posting fails.

#### Scenario: Thinking placeholder posted
- **WHEN** a user message is received in Telegram
- **THEN** the channel SHALL send a "_Thinking..._" message with Markdown parse mode

#### Scenario: Response delivered via edit
- **WHEN** the agent returns a successful response and a placeholder exists
- **THEN** the placeholder SHALL be edited with the response text

#### Scenario: Fallback to typing indicator
- **WHEN** posting the placeholder fails
- **THEN** the channel SHALL fall back to the existing typing indicator behavior

### Requirement: Discord progressive thinking indicator
The Discord channel SHALL post a "Thinking..." placeholder message and periodically edit it with elapsed time. It SHALL fall back to typing indicators if posting fails.

#### Scenario: Thinking placeholder posted
- **WHEN** a user message is received in Discord
- **THEN** the channel SHALL send a "_Thinking..._" message

#### Scenario: Response delivered via edit
- **WHEN** the agent returns a successful response and a placeholder exists
- **THEN** the placeholder SHALL be edited with the response content

#### Scenario: Long response truncated on edit
- **WHEN** the response exceeds Discord's 2000-character limit during edit
- **THEN** the content SHALL be truncated to 1997 characters plus "..."

### Requirement: Gateway progress broadcast
The gateway SHALL broadcast `agent.progress` events every 15 seconds during agent execution, including the elapsed time.

#### Scenario: Progress event broadcast
- **WHEN** 15 seconds have elapsed during agent execution
- **THEN** the gateway SHALL broadcast an `agent.progress` event with `elapsed` and `message` fields

#### Scenario: Progress stopped on completion
- **WHEN** the agent completes (success or error)
- **THEN** progress broadcasting SHALL stop

### Requirement: Gateway structured error event
The gateway SHALL broadcast `agent.error` events with structured fields including error code, user message, partial result, and hint.

#### Scenario: AgentError broadcast with full fields
- **WHEN** the agent returns an `AgentError` with code, partial, and user message
- **THEN** the `agent.error` event SHALL include `code`, `error` (user message), `partial`, and `hint` fields

#### Scenario: Plain error broadcast
- **WHEN** the agent returns a non-AgentError
- **THEN** the `agent.error` event SHALL include `error` with the raw message and empty `code`/`partial`/`hint`
