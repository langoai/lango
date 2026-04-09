## Purpose

Capability spec for tui-tool-lifecycle. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Tool transcript item kind
The chat transcript SHALL support an `itemTool` kind that represents a single tool invocation with lifecycle state.

#### Scenario: Tool call appears as distinct transcript item
- **WHEN** a `ToolStartedMsg` is received during streaming
- **THEN** a new transcript item of kind `itemTool` is appended with state `running`, displaying the tool name and an activity icon

#### Scenario: Tool completion updates item state
- **WHEN** a `ToolFinishedMsg` is received with `Success: true`
- **THEN** the corresponding tool item transitions to state `success` with duration displayed

#### Scenario: Tool failure updates item state
- **WHEN** a `ToolFinishedMsg` is received with `Success: false`
- **THEN** the corresponding tool item transitions to state `error` with the error output preview

### Requirement: Tool item state machine
Each tool transcript item SHALL maintain a state from the set: `running`, `success`, `error`, `canceled`, `awaiting_approval`. Transitions are one-directional from `running` to any terminal state.

#### Scenario: State transitions
- **WHEN** a tool item is in state `running`
- **THEN** it MAY transition to `success`, `error`, `canceled`, or `awaiting_approval` but SHALL NOT return to `running`

### Requirement: Tool item renderer
The tool renderer SHALL display a state-specific icon, tool name, and duration. Icons: ⚙ running, ✓ success, ✗ error, ⊘ canceled, 🔒 awaiting_approval.

#### Scenario: Running tool display
- **WHEN** a tool item is in state `running`
- **THEN** it renders as `⚙ [tool_name] running...` with muted accent

#### Scenario: Completed tool display
- **WHEN** a tool item is in state `success` with duration 2.3s
- **THEN** it renders as `✓ [tool_name] (2.3s)` with success accent and optional output preview

### Requirement: Tool item tracks callID
Each tool item SHALL be keyed by `callID` to correlate `ToolStartedMsg` with `ToolFinishedMsg`.

#### Scenario: Multiple concurrent tool calls
- **WHEN** two `ToolStartedMsg` arrive with different callIDs
- **THEN** two separate tool items are created and each finalizes independently
