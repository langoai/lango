## Purpose

Capability spec for tui-chat-rendering. See requirements below for scope and behavior contracts.
## Requirements
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

#### Scenario: Header and turn strip stay single-line on narrow terminals
- **WHEN** the chat header or turn status strip renders on a narrow terminal
- **THEN** each bar SHALL remain a single rendered line
- **AND** each bar SHALL clamp to the available terminal width instead of wrapping

#### Scenario: Header rendering fails closed when config is unavailable
- **WHEN** the chat header renders with a nil config pointer
- **THEN** it SHALL not panic
- **AND** it SHALL fall back to the existing default provider/model labels

#### Scenario: Header display fields stay plain and single-line
- **WHEN** provider, model, or session-key display text contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the chat header SHALL strip those control sequences
- **AND** it SHALL normalize the display text to a single line before rendering

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

#### Scenario: User transcript blocks stay plain visible text
- **WHEN** a user transcript block renders submitted content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before displaying the prompt text

#### Scenario: In-flight streaming rows stay plain visible text
- **WHEN** the live streaming transcript row renders chunk content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before displaying the partial assistant output

#### Scenario: Display-only transcript fields are stored in sanitized form
- **WHEN** the chat model appends `system`, `status`, `channel`, or `delegation` transcript data that contains ANSI/OSC escape sequences
- **THEN** the stored transcript fields used for display SHALL already be stripped and normalized rather than preserving raw control-sequence text

#### Scenario: Stored channel transcript payloads are sanitized
- **WHEN** the chat model appends a `channel` transcript row whose message text contains ANSI/OSC escape sequences
- **THEN** the stored transcript payload used for rendering SHALL already be stripped and normalized rather than preserving raw control-sequence text

#### Scenario: Stored recovery transcript cause metadata is sanitized
- **WHEN** the chat model appends a `recovery` transcript row whose `causeClass` contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the stored recovery metadata SHALL already be stripped and normalized rather than preserving raw control-sequence text

#### Scenario: Stored tool transcript names are sanitized
- **WHEN** the chat model appends a tool transcript row whose tool name contains ANSI/OSC escape sequences or embedded newlines
- **THEN** the stored transcript entry name SHALL already be stripped and normalized rather than preserving raw control-sequence text

#### Scenario: Status message rendered compactly
- **WHEN** the runtime emits a warning, cancel, or approval resolution message
- **THEN** the transcript SHALL render it as a compact `status` item instead of a full assistant prose block

#### Scenario: System transcript blocks stay plain visible text
- **WHEN** a `system` transcript block renders content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before displaying the block body

#### Scenario: Compact status and approval rows stay plain and single-line
- **WHEN** a compact `status` or `approval` transcript row renders content containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed content to a single line before truncating it into the row

#### Scenario: Stored status transcript content is normalized to single-line text
- **WHEN** the chat model appends a compact `status` transcript item whose content contains embedded newlines or terminal control sequences
- **THEN** the stored status content SHALL already be stripped and normalized to single-line text before rendering

#### Scenario: Approval transcript events stay plain and single-line
- **WHEN** an approval transcript event or its request-summary preview contains ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed event text to a single line before storing it in the transcript

#### Scenario: Approval request-id annotations stay plain and single-line
- **WHEN** an approval transcript event renders a request ID containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed request-id annotation to a single line before compacting it

#### Scenario: Thinking transcript rows stay compact and width-safe
- **WHEN** a `thinking` transcript row renders preview text on a narrow terminal
- **THEN** the row SHALL stay single-line
- **AND** the rendered row SHALL clamp to the available transcript width instead of overflowing

#### Scenario: Thinking transcript rows stay plain and single-line
- **WHEN** a `thinking` transcript row renders preview or fallback text containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before truncating it into the row

#### Scenario: Delegation transcript rows stay compact and width-safe
- **WHEN** a `delegation` transcript row renders long actor names or a multiline reason on a narrow terminal
- **THEN** the row SHALL stay single-line
- **AND** the rendered row SHALL clamp to the available transcript width instead of overflowing

#### Scenario: Delegation transcript rows sanitize actor names
- **WHEN** a `delegation` transcript row renders `from` or `to` names containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed actor names to a single line before rendering them

#### Scenario: Recovery transcript rows stay compact and single-line
- **WHEN** a `recovery` transcript row renders multiline or whitespace-heavy recovery metadata
- **THEN** the row SHALL normalize that metadata to a single line
- **AND** the rendered row SHALL remain a compact one-line event

#### Scenario: Recovery transcript rows keep cause metadata plain and single-line
- **WHEN** a `recovery` transcript row renders `causeClass` text containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed cause text to a single line before rendering it

#### Scenario: Recovery transcript rows keep action labels on the sanitized known-action path
- **WHEN** a `recovery` transcript row renders a known action value wrapped in ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences before mapping the action to its display label

#### Scenario: Channel transcript rows stay compact and width-safe
- **WHEN** a `channel` transcript row renders long sender names or multiline remote message text on a narrow terminal
- **THEN** the row SHALL normalize those fields to a single line
- **AND** the rendered row SHALL clamp to the available transcript width instead of overflowing

#### Scenario: Channel transcript rows sanitize remote sender and message text
- **WHEN** a `channel` transcript row renders remote sender or message text containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before display
- **AND** the rendered row SHALL remain a plain visible text surface

#### Scenario: Channel transcript rows sanitize displayed badge text
- **WHEN** a `channel` transcript row renders a channel name containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed badge text to a single line before rendering it

#### Scenario: Channel badge color follows the sanitized channel name
- **WHEN** a `channel` transcript row receives a known channel name wrapped in ANSI/OSC escape sequences
- **THEN** the badge text SHALL use the sanitized visible name
- **AND** the badge color selection SHALL use that same sanitized channel name rather than the raw control-sequence input

#### Scenario: Tool transcript rows keep detail previews width-safe
- **WHEN** a `tool` transcript row renders a preview or output/error detail line on a narrow terminal
- **THEN** each visible line SHALL clamp to the available transcript width instead of overflowing
- **AND** preview/output text SHALL be normalized to a single line before rendering

#### Scenario: Tool detail lines stay plain and single-line
- **WHEN** a tool preview or output/error detail line contains ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed detail text to a single line before truncating it into the row

#### Scenario: Tool and approval surfaces sanitize displayed tool names
- **WHEN** a chat approval surface or tool-lifecycle row renders a tool name containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed tool name to a single line before rendering it

#### Scenario: Parameter display surfaces sanitize displayed parameter keys
- **WHEN** a chat approval surface or tool param preview renders parameter keys containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed key text to a single line before rendering it

#### Scenario: Pending indicator stays width-safe
- **WHEN** the submit-to-first-event pending indicator renders on a narrow terminal
- **THEN** it SHALL clamp to the available transcript width instead of overflowing

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

### Requirement: Markdown rendering fails closed to plain text
The chat TUI SHALL preserve transcript stability when the markdown renderer errors or panics.

#### Scenario: Renderer error falls back to plain text
- **WHEN** `renderMarkdown()` cannot render through Glamour because the renderer returns an error
- **THEN** it SHALL return the original content as plain text instead of failing the transcript

#### Scenario: Renderer panic falls back to plain text
- **WHEN** `renderMarkdown()` encounters a renderer panic
- **THEN** it SHALL recover and return the original content as plain text instead of crashing the TUI

#### Scenario: Sanitized markdown input is used for rendering
- **WHEN** assistant markdown input contains ANSI/OSC escape sequences
- **THEN** the chat TUI SHALL strip those control sequences before rendering the markdown

#### Scenario: Sanitized markdown input is used for plain-text fallback
- **WHEN** assistant markdown rendering fails after receiving input that contained ANSI/OSC escape sequences
- **THEN** the fallback plain-text transcript content SHALL use the sanitized text rather than the raw control-sequence input

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

#### Scenario: Stored assistant raw markdown strips control sequences
- **WHEN** assistant markdown input contains ANSI/OSC escape sequences
- **THEN** the stored raw markdown used for reflow SHALL strip those control sequences while preserving the remaining markdown/newline structure

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

#### Scenario: Duplicate error suppression compares sanitized assistant text
- **WHEN** DoneMsg arrives with a non-success outcome and ResponseText differs from the last assistant raw content only by stripped control sequences
- **THEN** the duplicate status/error entry SHALL still be skipped

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

#### Scenario: Idle and failed help describe double-press quit
- **WHEN** the chat help bar is rendered in the `idle` or `failed` states
- **THEN** the `Ctrl+C` binding SHALL describe quitting via the double-press path rather than immediate single-press exit

#### Scenario: Idle and failed help advertise immediate Ctrl+D quit
- **WHEN** the chat help bar is rendered in the `idle` or `failed` states
- **THEN** it SHALL advertise `Ctrl+D` as the immediate quit path

#### Scenario: Chat help uses unified transcript-scroll wording
- **WHEN** a user reads chat help for `PgUp/PgDn`
- **THEN** the scrolling action SHALL be described with one consistent transcript-scroll phrase across in-product and public docs

#### Scenario: Approval surfaces use unified session-allow wording
- **WHEN** a chat approval action bar is rendered for either inline or fullscreen approval UI
- **THEN** the `s` action SHALL be labeled `allow session`

#### Scenario: Approval surfaces use consistent deny-key wording
- **WHEN** a chat approval surface renders a deny affordance
- **THEN** it SHALL label the deny keys consistently as `d/Esc`

#### Scenario: Approval-state turn strip uses unified deny-key wording
- **WHEN** the turn status strip renders in the `approving` state
- **THEN** it SHALL surface the deny path using the shared `d/Esc` wording

#### Scenario: Inline approval strip keeps summary plain and single-line
- **WHEN** the Tier 1 inline approval strip renders a summary containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the summary to a single line before truncating it into the strip

#### Scenario: Approval surfaces keep summaries plain and single-line
- **WHEN** any chat approval surface renders a summary containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the summary to a single line before displaying it

#### Scenario: Approval surfaces sanitize displayed channel origin text
- **WHEN** a chat approval surface renders channel-origin text extracted from a session key that contains ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed origin text to a single line before rendering it

#### Scenario: Approval origin/badge matching uses the sanitized session-key prefix
- **WHEN** an approval surface receives a known channel prefix wrapped in ANSI/OSC escape sequences inside the session key
- **THEN** it SHALL strip those control sequences before matching the prefix to Telegram, Discord, or Slack

#### Scenario: Fullscreen approval dialog keeps risk text plain and single-line
- **WHEN** the fullscreen approval dialog renders risk-label or rule-explanation text containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed risk text to plain single-line text before rendering it

#### Scenario: Fullscreen approval dialog keeps risk badge text plain and single-line
- **WHEN** the fullscreen approval dialog renders a risk level containing ANSI/OSC escape sequences or embedded newlines
- **THEN** it SHALL strip those control sequences
- **AND** it SHALL normalize the displayed badge text to plain single-line text before rendering it

#### Scenario: Fullscreen approval badge color follows the sanitized risk level
- **WHEN** the fullscreen approval dialog receives a known risk level wrapped in ANSI/OSC escape sequences
- **THEN** the badge text SHALL use the sanitized visible level
- **AND** the badge color selection SHALL use that same sanitized risk level rather than the raw control-sequence input

#### Scenario: Fullscreen approval dialog sanitizes diff preview text
- **WHEN** the fullscreen approval dialog renders diff content containing ANSI/OSC escape sequences
- **THEN** it SHALL strip those control sequences before styling or displaying the diff lines

#### Scenario: Approval confirm prompt reflects the pending action key
- **WHEN** a critical-risk approval surface is in confirm-pending state
- **THEN** the visible confirm prompt SHALL name the actual pending action key (`a` or `s`) rather than a hard-coded default

#### Scenario: Approval confirm prompt keeps deny path visible
- **WHEN** a critical-risk approval surface is in confirm-pending state
- **THEN** the visible confirm prompt SHALL mention that `d` or `Esc` still denies the request

#### Scenario: Fullscreen approval dialog surfaces split-toggle help when diff exists
- **WHEN** the fullscreen approval dialog renders diff content
- **THEN** its action bar SHALL advertise the `t` key for toggling split mode

#### Scenario: Fullscreen approval dialog shows scroll help only when diff overflow exists
- **WHEN** the fullscreen approval dialog renders diff content that exceeds the visible diff area
- **THEN** its action bar SHALL advertise `↑/k` and `↓/j` scrolling

#### Scenario: Fullscreen approval dialog hides inert scroll help when diff fits
- **WHEN** the fullscreen approval dialog renders diff content that fits within the visible diff area
- **THEN** its action bar SHALL omit `↑/k` and `↓/j` scrolling

#### Scenario: Fullscreen approval diff clamps scroll offset when the diff fits
- **WHEN** the fullscreen approval dialog renders a diff that fits within the visible diff area
- **THEN** its scroll offset SHALL clamp to zero
- **AND** the full short diff SHALL remain visible

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
