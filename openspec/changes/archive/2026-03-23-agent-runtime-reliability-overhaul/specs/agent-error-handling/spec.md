## ADDED Requirements

### Requirement: Empty-after-tool-use error classification
If an agent run terminates with no visible assistant completion after one or more successful tool results, the runtime SHALL return a dedicated structured agent error classification for that condition rather than reporting a silent success.

#### Scenario: Channel path receives classified empty-after-tool-use error
- **WHEN** `RunAndCollect` finishes without visible text after successful specialist tool results
- **THEN** the runtime SHALL return a structured agent error classified as `empty_after_tool_use`
- **AND** the channel path SHALL surface the corresponding user-facing message instead of relying on the generic empty-success fallback

#### Scenario: Gateway path broadcasts classified empty-after-tool-use error
- **WHEN** `RunStreaming` finishes without visible text after successful specialist tool results
- **THEN** the gateway path SHALL emit an `agent.error` event carrying the `empty_after_tool_use` classification
- **AND** SHALL NOT treat the turn as a successful empty response

### Requirement: Call-signature loop classification
The runtime SHALL classify repeated same-agent same-tool same-params sequences as loop failures even when call IDs differ or tool response events are interleaved between attempts.

#### Scenario: Different call IDs do not bypass loop classification
- **WHEN** the same agent repeatedly emits the same tool name with canonically equal params but different generated call IDs
- **THEN** the runtime SHALL still classify the sequence as the same loop signature
- **AND** SHALL terminate the run with `loop_detected` once the configured threshold is exceeded

#### Scenario: Tool response interleaving does not reset the loop signature
- **WHEN** a specialist alternates `FunctionCall` and `FunctionResponse` events for the same canonical tool signature without visible assistant progress
- **THEN** the runtime SHALL continue counting the repeated signature
- **AND** SHALL NOT reset the loop counter solely because tool responses were observed

### Requirement: Truthful recovery messaging
User-facing recovery messages after loop, timeout, or empty-after-tool-use failures SHALL only describe evidence actually gathered during the current turn. They SHALL NOT claim that unavailable tools were directly executed.

#### Scenario: Recovery message avoids unavailable direct-call claim
- **WHEN** the orchestrator has no direct access to `payment_balance` and recovery messaging is generated after a specialist failure
- **THEN** the user-facing message SHALL NOT claim that the orchestrator directly executed `payment_balance`
- **AND** SHALL instead describe the actual specialist failure or previously gathered evidence truthfully
