### Requirement: Parts-based layout agreement
The `View()` method and `recalcLayout()` method SHALL use the same parts structure so that measured heights always match rendered output. The viewport height SHALL be computed by subtracting the measured heights of all fixed parts (header, turn status strip, composer or approval card, help footer) and separators from the terminal height.

#### Scenario: Layout matches rendered output
- **WHEN** the terminal is 80x24
- **THEN** the sum of all rendered fixed part heights, separator newlines, and viewport height SHALL fit within the terminal height

#### Scenario: Minimum viewport height
- **WHEN** the terminal height is very small (e.g., height=5)
- **THEN** the viewport height SHALL be clamped to a minimum of 3

#### Scenario: Approval state recalculates layout
- **WHEN** an ApprovalRequestMsg is received
- **THEN** recalcLayout() SHALL be called so the approval card height replaces the composer height in the layout calculation

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

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: User message rendered as transcript item
- **WHEN** the user submits a prompt
- **THEN** the transcript SHALL add a `user` item rendered with the user block style

#### Scenario: Status message rendered compactly
- **WHEN** the runtime emits a warning, cancel, or approval resolution message
- **THEN** the transcript SHALL render it as a compact `status` item instead of a full assistant prose block

### Requirement: Block-joined transcript spacing
The transcript viewport SHALL render message blocks by joining blocks explicitly rather than prefixing each block with leading newlines.

#### Scenario: No leading blank lines
- **WHEN** the transcript contains one or more items
- **THEN** the rendered viewport content SHALL NOT start with blank lines

#### Scenario: Stable spacing between blocks
- **WHEN** adjacent transcript items are rendered
- **THEN** they SHALL be separated by a consistent explicit gap rather than accumulating extra blank lines

### Requirement: Assistant append unification
All assistant-visible response content SHALL be created through a single append helper that stores raw markdown and computes rendered content for the current transcript content width.

#### Scenario: Stream finalization uses append helper
- **WHEN** streaming completes and buffered chunks exist
- **THEN** the transcript SHALL create one assistant item through the shared append helper

#### Scenario: Non-streaming fallback uses append helper
- **WHEN** a turn completes without buffered chunks but with non-empty ResponseText
- **THEN** the transcript SHALL create one assistant item through the same append helper

#### Scenario: Partial output preserved on cancel
- **WHEN** generation is cancelled after some streamed chunks were already received
- **THEN** the buffered content SHALL still be committed as an assistant item through the shared append helper

### Requirement: Markdown rendering performance
The chat TUI SHALL cache the glamour `TermRenderer` at module level, keyed by terminal width. The renderer SHALL be reused across `renderMarkdown()` calls at the same width. A new renderer SHALL only be created when width changes.

#### Scenario: Renderer reused on cursor tick
- **WHEN** `renderMarkdown` is called multiple times at the same width (e.g., cursor blink every 400ms)
- **THEN** the same cached renderer SHALL be reused without creating a new one

#### Scenario: Renderer recreated on width change
- **WHEN** the terminal width changes
- **THEN** a new renderer SHALL be created and cached for the new width

### Requirement: Transcript render optimization
The chat `render()` method SHALL use the pre-rendered `content` field for finalized assistant entries. It SHALL NOT re-invoke `renderMarkdown()` on every render pass. Re-rendering of assistant entries SHALL only occur in `setSize()` when the width actually changes.

#### Scenario: Cursor tick does not re-render finalized entries
- **WHEN** a cursor blink tick fires
- **THEN** `render()` SHALL use cached `entry.content` for all finalized assistant entries

#### Scenario: Width change triggers assistant re-render
- **WHEN** `setSize()` is called and width differs from previous
- **THEN** all assistant entries with `rawContent` SHALL have their `content` field re-rendered

#### Scenario: Height-only change skips re-render
- **WHEN** `setSize()` is called with the same width but different height
- **THEN** assistant entries SHALL NOT be re-rendered

### Requirement: Assistant raw markdown reflow
Assistant transcript items SHALL preserve raw markdown for re-rendering when the viewport width changes.

#### Scenario: Assistant raw markdown stored
- **WHEN** an assistant item is appended
- **THEN** the original markdown SHALL be stored in a raw content field in addition to the rendered display content

#### Scenario: Resize reflows assistant content
- **WHEN** the terminal width changes after assistant content has been rendered
- **THEN** assistant items SHALL be re-rendered from raw markdown using the current transcript content width

### Requirement: DoneMsg three-rule processing
DoneMsg SHALL be processed with three rules in order:
1. If streamBuf is non-empty, finalize it as an assistant message.
2. Else if ResponseText is non-empty, add it via appendAssistant.
3. If outcome is not "success", add a compact status or error entry with deduplication.

#### Scenario: Stream success
- **WHEN** DoneMsg arrives with outcome="success" and streamBuf has content
- **THEN** streamBuf SHALL be finalized as an assistant entry with rawContent preserved

#### Scenario: Non-streaming model response
- **WHEN** DoneMsg arrives with empty streamBuf but non-empty ResponseText
- **THEN** ResponseText SHALL be added as an assistant entry via appendAssistant

#### Scenario: Failure preserves partial stream
- **WHEN** DoneMsg arrives with outcome="timeout" and streamBuf has content
- **THEN** the partial stream SHALL be finalized as an assistant entry AND a compact status/error entry SHALL be added

#### Scenario: Duplicate error text suppression
- **WHEN** DoneMsg arrives with non-success outcome and ResponseText matches the last assistant rawContent
- **THEN** the duplicate status/error entry SHALL be skipped

### Requirement: ErrorMsg partial-first preservation
When an ErrorMsg is received, any in-flight stream content SHALL be finalized as an assistant message before a status or error entry is added.

#### Scenario: Error with partial stream
- **WHEN** ErrorMsg arrives while streamBuf has content
- **THEN** the stream content SHALL be preserved as an assistant entry AND an error status entry SHALL be added

#### Scenario: Cancel returns to idle
- **WHEN** ErrorMsg arrives with `context.Canceled`
- **THEN** the TUI SHALL preserve any partial stream content, append a cancellation status entry, and return to idle state

### Requirement: Turn state strip
The TUI SHALL render a dedicated turn status strip that reflects at least the states `idle`, `streaming`, `approving`, `cancelling`, and `failed`.

#### Scenario: Streaming state visible
- **WHEN** the agent begins generating a response
- **THEN** the turn status strip SHALL show that generation is in progress and cancellation is available

#### Scenario: Approval state visible
- **WHEN** a tool approval request interrupts the current turn
- **THEN** the turn status strip SHALL show that approval is required

#### Scenario: Failed state visible
- **WHEN** a turn ends in failure without producing a successful completion
- **THEN** the turn status strip SHALL show a failed state until the next user interaction resets it

### Requirement: Composer remains visible during streaming
During streaming, the composer SHALL remain visible in a read-only or visually muted state instead of being removed from the layout.

#### Scenario: Streaming keeps composer visible
- **WHEN** the TUI enters streaming state
- **THEN** the composer SHALL remain visible and indicate that input is temporarily unavailable

#### Scenario: Approval hides composer
- **WHEN** the TUI enters approval state
- **THEN** the composer SHALL be replaced by the approval card for the duration of the approval interruption

### Requirement: Approval banner width clamp
The `renderApprovalBanner()` function SHALL clamp the banner width to `max(width - 4, 10)` to prevent layout issues at narrow terminal widths.

#### Scenario: Normal width
- **WHEN** terminal width is 80
- **THEN** banner content width SHALL be 76

#### Scenario: Narrow terminal
- **WHEN** terminal width is 8
- **THEN** banner content width SHALL be clamped to 10

### Requirement: Content width for markdown rendering
The transcript content width helper SHALL return `max(width - 2, 10)` as the available width for assistant markdown rendering, accounting for left indent and safety margin.

#### Scenario: Standard width
- **WHEN** viewport width is 80
- **THEN** the transcript content width SHALL be 78

#### Scenario: Minimum clamp
- **WHEN** viewport width is 5
- **THEN** the transcript content width SHALL be clamped to 10

### Requirement: Mouse wheel scrolling support
The bubbletea program SHALL be created with `tea.WithMouseCellMotion()` to enable mouse event delivery. The viewport SHALL receive mouse wheel events for scrolling through chat history.

#### Scenario: Mouse wheel scrolls viewport
- **WHEN** the user scrolls the mouse wheel over the chat viewport
- **THEN** the viewport content SHALL scroll accordingly

#### Scenario: No hover event noise
- **WHEN** the user moves the mouse without clicking or scrolling
- **THEN** no mouse motion events SHALL be delivered

### Requirement: TUI log file redirect
In TUI chat mode, logging SHALL be redirected to a file at `<DataRoot>/chat.log` instead of stderr. The log file path SHALL be displayed to the user during TUI initialization.

#### Scenario: No log corruption on screen
- **WHEN** async goroutines emit logs during TUI operation
- **THEN** the log output SHALL NOT appear on the alt-screen TUI display

#### Scenario: Log file path displayed
- **WHEN** the TUI starts
- **THEN** the log file path SHALL be printed to stderr before entering alt-screen mode

#### Scenario: Logs written to file
- **WHEN** any component writes log output during a TUI session
- **THEN** the log entry SHALL be appended to `<DataRoot>/chat.log`
