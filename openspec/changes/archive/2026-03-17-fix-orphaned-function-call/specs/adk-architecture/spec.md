## MODIFIED Requirements

### Requirement: FunctionResponse events are stored with correct role

The system SHALL correct the role of FunctionResponse events from `"user"` to `"tool"` at write-time in `AppendEvent`. A message is classified as FunctionResponse-only when it contains ToolCalls with `Output != ""` and no ToolCalls with `Input != ""`.

#### Scenario: ADK sends FunctionResponse with role "user"
- **WHEN** `AppendEvent` receives an event with `Content.Role = "user"` containing a FunctionResponse part
- **THEN** the persisted message SHALL have `Role = "tool"`

#### Scenario: FunctionCall event role is unchanged
- **WHEN** `AppendEvent` receives an event with `Content.Role = "model"` containing a FunctionCall part
- **THEN** the persisted message SHALL have `Role = "assistant"` (normalized from "model")

### Requirement: Legacy FunctionResponse data is corrected at read-time

The system SHALL correct the role of FunctionResponse messages stored with `Role = "user"` during `EventsAdapter.All()` reconstruction. Messages with ToolCalls containing `Output != ""` and stored role `"user"` SHALL be treated as role `"tool"` for event reconstruction purposes.

#### Scenario: FunctionResponse stored as "user" is reconstructed correctly
- **WHEN** `EventsAdapter.All()` encounters a message with `Role = "user"` and ToolCalls containing `Output != ""`
- **THEN** the reconstructed event SHALL have `Content.Role = "function"` and contain `FunctionResponse` parts
- **AND** the event author SHALL be `"tool"`

#### Scenario: Correctly stored FunctionResponse is unaffected
- **WHEN** `EventsAdapter.All()` encounters a message with `Role = "tool"` and ToolCalls containing `Output != ""`
- **THEN** the reconstructed event SHALL have `Content.Role = "function"` and contain `FunctionResponse` parts (unchanged behavior)

### Requirement: Orphaned FunctionCalls are repaired at provider boundary

The system SHALL inject synthetic error tool responses for orphaned FunctionCalls in `convertMessages` when an assistant message with FunctionCalls is followed by a user message without intervening tool responses for all calls. Pending FunctionCalls at the end of history SHALL NOT be modified.

#### Scenario: Orphaned FunctionCall followed by user message
- **WHEN** `convertMessages` encounters an assistant message with a FunctionCall followed by a user message with no intervening tool response
- **THEN** the system SHALL inject a synthetic tool response with error content before the user message
- **AND** SHALL log a WARN-level message with the call ID

#### Scenario: Matched FunctionCall with tool response
- **WHEN** `convertMessages` encounters an assistant FunctionCall followed by a matching tool response and then a user message
- **THEN** no synthetic response SHALL be injected

#### Scenario: Pending FunctionCall at end of history
- **WHEN** `convertMessages` encounters an assistant FunctionCall as the last message (or no user message follows)
- **THEN** no synthetic response SHALL be injected

#### Scenario: Partially answered FunctionCalls
- **WHEN** an assistant message contains multiple FunctionCalls and only some have matching tool responses before the next user message
- **THEN** synthetic responses SHALL be injected only for the unanswered calls
