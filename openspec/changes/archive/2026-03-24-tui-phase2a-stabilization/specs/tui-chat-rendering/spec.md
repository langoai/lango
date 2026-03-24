## MODIFIED Requirements

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

## ADDED Requirements

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
