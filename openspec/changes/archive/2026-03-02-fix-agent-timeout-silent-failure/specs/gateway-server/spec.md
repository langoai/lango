## MODIFIED Requirements

### Requirement: Agent thinking events
The Gateway server SHALL broadcast `agent.thinking` before agent processing. On successful completion, it SHALL broadcast `agent.done`. On error, it SHALL broadcast `agent.error` with error details and classification. The server SHALL also broadcast `agent.warning` when approaching the request timeout (80% elapsed).

#### Scenario: Thinking event on message receipt
- **WHEN** a `chat.message` RPC is received
- **THEN** the server SHALL broadcast an `agent.thinking` event to the session before calling `RunStreaming`

#### Scenario: Done event after successful processing
- **WHEN** `RunStreaming` returns successfully
- **THEN** the server SHALL broadcast an `agent.done` event to the session

#### Scenario: Error event after failed processing
- **WHEN** `RunStreaming` returns an error
- **THEN** the server SHALL broadcast an `agent.error` event to the session
- **AND** the event payload SHALL include `error` (error message string) and `type` (error classification)
- **AND** the `type` SHALL be `"timeout"` when `ctx.Err() == context.DeadlineExceeded`, otherwise `"unknown"`
- **AND** `agent.done` SHALL NOT be broadcast

#### Scenario: Warning event when approaching timeout
- **WHEN** 80% of the request timeout duration has elapsed
- **AND** the agent is still processing
- **THEN** the server SHALL broadcast an `agent.warning` event to the session
- **AND** the event payload SHALL include `type: "approaching_timeout"` and a human-readable `message`
