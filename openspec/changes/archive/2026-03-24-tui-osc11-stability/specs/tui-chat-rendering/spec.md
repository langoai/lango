## MODIFIED Requirements

### Requirement: Typed transcript rendering
The transcript viewport SHALL render typed transcript items rather than plain role/content rows. The minimum item kinds SHALL be `user`, `assistant`, `system`, `status`, and `approval`.

#### Scenario: User message rendered as transcript item
- **WHEN** the user submits a prompt
- **THEN** the transcript SHALL add a `user` item rendered with the user block style

#### Scenario: Status message rendered compactly
- **WHEN** the runtime emits a warning, cancel, or approval resolution message
- **THEN** the transcript SHALL render it as a compact `status` item instead of a full assistant prose block

#### Scenario: User and assistant visually distinct
- **WHEN** adjacent user and assistant items are rendered in the transcript
- **THEN** their blocks SHALL use visibly different accents, tinting, or separators so they can be distinguished at a glance

### Requirement: Assistant raw markdown reflow
Assistant transcript items SHALL preserve raw markdown for re-rendering when the viewport width changes.

#### Scenario: Assistant raw markdown stored
- **WHEN** an assistant item is appended
- **THEN** the original markdown SHALL be stored in a raw content field in addition to the rendered display content

#### Scenario: Resize reflows assistant content
- **WHEN** the terminal width changes after assistant content has been rendered
- **THEN** assistant items SHALL be re-rendered from raw markdown using the current transcript content width

### Requirement: Content width for markdown rendering
The transcript content width helper SHALL return `max(width - 2, 10)` as the available width for assistant markdown rendering, accounting for left indent and safety margin.

#### Scenario: Standard width
- **WHEN** viewport width is 80
- **THEN** the transcript content width SHALL be 78

#### Scenario: Minimum clamp
- **WHEN** viewport width is 5
- **THEN** the transcript content width SHALL be clamped to 10

### Requirement: Fixed markdown style in TUI mode
The TUI markdown renderer SHALL use an explicit Glamour standard style instead of auto style detection.

#### Scenario: No terminal background query for markdown rendering
- **WHEN** assistant markdown is rendered in TUI mode
- **THEN** the renderer SHALL NOT query terminal background color through auto-style detection

#### Scenario: Dark style default
- **WHEN** TUI markdown rendering is initialized without an explicit user theme override
- **THEN** Glamour dark style SHALL be used as the default renderer style
