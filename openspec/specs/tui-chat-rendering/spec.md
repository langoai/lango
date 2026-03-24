### Requirement: Parts-based layout agreement
The `View()` method and `recalcLayout()` method SHALL use the same parts structure so that measured heights always match rendered output. The viewport height SHALL be computed by subtracting the measured heights of all fixed parts (status bar, input/approval banner, help bar) and separators from the terminal height.

#### Scenario: Layout matches rendered output
- **WHEN** the terminal is 80x24
- **THEN** the sum of all rendered part heights plus separator newlines equals the terminal height

#### Scenario: Minimum viewport height
- **WHEN** the terminal height is very small (e.g., height=5)
- **THEN** the viewport height SHALL be clamped to a minimum of 3

#### Scenario: Approval state recalculates layout
- **WHEN** an ApprovalRequestMsg is received
- **THEN** recalcLayout() SHALL be called to reflect the approval banner height instead of input height

### Requirement: Input width safety margin
The input component SHALL set the textarea width to `max(terminalWidth - 2, 10)` to account for border padding and prevent border wrapping.

#### Scenario: Normal terminal width
- **WHEN** terminal width is 80
- **THEN** textarea width SHALL be set to 78

#### Scenario: Very narrow terminal
- **WHEN** terminal width is 8
- **THEN** textarea width SHALL be clamped to minimum 10

#### Scenario: No border triplication
- **WHEN** the input is rendered at any terminal width
- **THEN** no input line SHALL exceed the terminal width

### Requirement: Block-join chat rendering
The chat viewport `render()` method SHALL collect message entries into discrete blocks and join them with `"\n\n"`. The rendered output SHALL NOT start with leading blank lines.

#### Scenario: No leading blank lines
- **WHEN** the chat has one or more entries
- **THEN** the viewport content SHALL NOT start with `"\n\n"`

#### Scenario: Consistent inter-block spacing
- **WHEN** multiple messages are rendered
- **THEN** each pair of adjacent messages SHALL be separated by exactly one blank line (`"\n\n"` join)

### Requirement: Assistant rawContent preservation
Every assistant entry SHALL store the original markdown in `rawContent` for resize reflow. The `appendAssistant(raw)` helper SHALL be the single entry point for all assistant message creation.

#### Scenario: Stream finalization preserves raw
- **WHEN** streaming completes and `finalizeStream()` is called
- **THEN** the resulting entry SHALL have `rawContent` equal to the original stream buffer content

#### Scenario: Non-streaming response preserves raw
- **WHEN** a DoneMsg arrives with ResponseText but no stream chunks
- **THEN** the resulting entry SHALL have `rawContent` equal to ResponseText

#### Scenario: Resize reflow
- **WHEN** the terminal is resized to a different width
- **THEN** assistant entries SHALL be re-rendered from `rawContent` using the new `contentWidth()`

### Requirement: DoneMsg three-rule processing
DoneMsg SHALL be processed with three rules in order:
1. If streamBuf is non-empty, finalize it as an assistant message.
2. Else if ResponseText is non-empty, add it via appendAssistant.
3. If outcome is not "success", add a system error message with deduplication.

#### Scenario: Stream success
- **WHEN** DoneMsg arrives with outcome="success" and streamBuf has content
- **THEN** streamBuf SHALL be finalized as an assistant entry with rawContent preserved

#### Scenario: Non-streaming model response
- **WHEN** DoneMsg arrives with empty streamBuf but non-empty ResponseText
- **THEN** ResponseText SHALL be added as an assistant entry via appendAssistant

#### Scenario: Failure preserves partial stream
- **WHEN** DoneMsg arrives with outcome="timeout" and streamBuf has content
- **THEN** the partial stream SHALL be finalized as an assistant entry AND a system error message SHALL be added

#### Scenario: Duplicate error text suppression
- **WHEN** DoneMsg arrives with non-success outcome and ResponseText matches the last assistant rawContent
- **THEN** the system error message SHALL be skipped to avoid duplication

### Requirement: ErrorMsg partial-first preservation
When an ErrorMsg is received, any in-flight stream content SHALL be finalized as an assistant message before the error is added as a system message.

#### Scenario: Error with partial stream
- **WHEN** ErrorMsg arrives while streamBuf has content
- **THEN** the stream content SHALL be preserved as an assistant entry AND the error SHALL be added as a separate system entry

#### Scenario: Error without stream
- **WHEN** ErrorMsg arrives with empty streamBuf
- **THEN** only the error system message SHALL be added

### Requirement: Approval banner width clamp
The `renderApprovalBanner()` function SHALL clamp the banner width to `max(width - 4, 10)` to prevent layout issues at narrow terminal widths.

#### Scenario: Normal width
- **WHEN** terminal width is 80
- **THEN** banner content width SHALL be 76

#### Scenario: Narrow terminal
- **WHEN** terminal width is 8
- **THEN** banner content width SHALL be clamped to 10

### Requirement: Content width for markdown rendering
The `contentWidth()` method SHALL return `max(width - 2, 10)` as the available width for assistant markdown rendering, accounting for left indent and safety margin.

#### Scenario: Standard width
- **WHEN** viewport width is 80
- **THEN** contentWidth() SHALL return 78

#### Scenario: Minimum clamp
- **WHEN** viewport width is 5
- **THEN** contentWidth() SHALL return 10

### Requirement: Mouse wheel scrolling support
The bubbletea program SHALL be created with `tea.WithMouseCellMotion()` to enable mouse event delivery. The viewport SHALL receive mouse wheel events for scrolling through chat history.

#### Scenario: Mouse wheel scrolls viewport
- **WHEN** the user scrolls the mouse wheel over the chat viewport
- **THEN** the viewport content SHALL scroll accordingly (up for wheel-up, down for wheel-down)

#### Scenario: No hover event noise
- **WHEN** the user moves the mouse without clicking or scrolling
- **THEN** no mouse motion events SHALL be delivered (WithMouseCellMotion, not WithMouseAllMotion)

### Requirement: TUI log file redirect
In TUI chat mode, logging SHALL be redirected to a file at `<DataRoot>/chat.log` instead of stderr. The log file path SHALL be displayed to the user during TUI initialization.

#### Scenario: No log corruption on screen
- **WHEN** async goroutines emit WARN or INFO logs during TUI operation
- **THEN** the log output SHALL NOT appear on the alt-screen TUI display

#### Scenario: Log file path displayed
- **WHEN** the TUI starts and displays the initialization banner
- **THEN** the log file path SHALL be printed to stderr before entering alt-screen mode

#### Scenario: Logs written to file
- **WHEN** any component writes log output during a TUI session
- **THEN** the log entry SHALL be appended to `<DataRoot>/chat.log`
