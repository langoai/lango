## Purpose

Capability spec for tui-thinking-indicator. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Thinking transcript item kind
The chat transcript SHALL support an `itemThinking` kind that represents an agent's reasoning/thinking phase, detected via `genai.Part.Thought == true`.

#### Scenario: Thinking block appears in transcript
- **WHEN** a `ThinkingStartedMsg` is received
- **THEN** a new transcript item of kind `itemThinking` is appended with a placeholder display

#### Scenario: Thinking block finalized with summary
- **WHEN** a `ThinkingFinishedMsg` is received with summary text and duration
- **THEN** the thinking item updates to show the summary and duration in a collapsible block

### Requirement: Thinking renderer
The thinking renderer SHALL display thinking blocks as collapsible items, defaulting to collapsed state showing duration only. The active state SHALL show "💭 Thinking..." followed by a truncated preview of the thinking summary text using `ansi.Truncate`. If the summary is empty, only "💭 Thinking..." is displayed.

#### Scenario: Active thinking with summary
- **WHEN** a ThinkingStartedMsg arrives with summary="analyzing user query for search terms"
- **THEN** the indicator SHALL display "💭 Thinking..." followed by a truncated preview of the summary in italic muted style

#### Scenario: Active thinking with empty summary
- **WHEN** a ThinkingStartedMsg arrives with summary=""
- **THEN** the indicator SHALL display "💭 Thinking..." with no preview text

#### Scenario: Summary truncated for narrow width
- **WHEN** the thinking summary exceeds the available display width
- **THEN** the preview SHALL be truncated with "..." using `ansi.Truncate` with a minimum preview width of 10

#### Scenario: Collapsed thinking display
- **WHEN** a finalized thinking item is rendered in collapsed state
- **THEN** it displays as a single line: `💭 Thinking (3.2s)` with muted accent

#### Scenario: Expanded thinking display
- **WHEN** a thinking item is expanded
- **THEN** it displays the full thinking text in a dimmed bordered block with duration header

### Requirement: Pending indicator for submit-to-first-event gap
The TUI SHALL display a `⏳ Working...` indicator from the moment a turn is submitted until the first chunk, tool event, or thinking event arrives. This covers responses that do not start with thinking.

#### Scenario: Pending indicator on submit
- **WHEN** a user submits a message and no events have arrived yet
- **THEN** the TUI displays `⏳ Working...` with elapsed time

#### Scenario: Pending indicator dismissed on first event
- **WHEN** the first `ChunkMsg`, `ToolStartedMsg`, or `ThinkingStartedMsg` arrives
- **THEN** the pending indicator is removed

### Requirement: Thinking detection from ADK events
The system SHALL detect thinking by checking `genai.Part.Thought == true` on parts within `session.Event.Content.Parts` during the turnrunner event loop.

#### Scenario: Thought part fires OnThinking callback
- **WHEN** `recordEvent()` encounters a part with `Thought == true` and non-empty `Text`
- **THEN** the `OnThinking` callback is fired with `started: true` and the thought text
