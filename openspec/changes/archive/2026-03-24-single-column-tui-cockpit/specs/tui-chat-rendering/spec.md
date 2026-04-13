## MODIFIED Requirements

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

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: User message rendered as transcript item
- **WHEN** the user submits a prompt
- **THEN** the transcript SHALL add a `user` item rendered with the user block style

#### Scenario: Status message rendered compactly
- **WHEN** the runtime emits a warning, cancel, or approval resolution message
- **THEN** the transcript SHALL render it as a compact `status` item instead of a full assistant prose block

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

### Requirement: Assistant raw markdown reflow
Assistant transcript items SHALL preserve raw markdown for re-rendering when the viewport width changes.

#### Scenario: Assistant raw markdown stored
- **WHEN** an assistant item is appended
- **THEN** the original markdown SHALL be stored in a raw content field in addition to the rendered display content

#### Scenario: Resize reflows assistant content
- **WHEN** the terminal width changes after assistant content has been rendered
- **THEN** assistant items SHALL be re-rendered from raw markdown using the current transcript content width

### Requirement: Block-joined transcript spacing
The transcript viewport SHALL render message blocks by joining blocks explicitly rather than prefixing each block with leading newlines.

#### Scenario: No leading blank lines
- **WHEN** the transcript contains one or more items
- **THEN** the rendered viewport content SHALL NOT start with blank lines

#### Scenario: Stable spacing between blocks
- **WHEN** adjacent transcript items are rendered
- **THEN** they SHALL be separated by a consistent explicit gap rather than accumulating extra blank lines

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
