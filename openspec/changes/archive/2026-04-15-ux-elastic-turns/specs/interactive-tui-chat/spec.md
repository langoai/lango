## MODIFIED Requirements

### Requirement: Real-time streaming responses
The TUI SHALL stream agent responses in real-time via `TurnRunner.Run()`. During streaming, the input composer SHALL remain focused and accept user input. If the user submits input during streaming, the current turn SHALL be cancelled and the new input SHALL be queued for immediate submission after the cancelled turn completes.

#### Scenario: Streaming output displayed incrementally
- **WHEN** the agent generates a response
- **THEN** text appears incrementally in the chat viewport as tokens arrive

#### Scenario: User can type during streaming
- **WHEN** the agent is streaming a response
- **THEN** the input composer SHALL be focused and accept text input

#### Scenario: User submits input during streaming
- **WHEN** the user presses Enter with non-empty input during streaming
- **THEN** the current turn SHALL be cancelled
- **AND** the new input SHALL be submitted as the next turn after cancellation completes
