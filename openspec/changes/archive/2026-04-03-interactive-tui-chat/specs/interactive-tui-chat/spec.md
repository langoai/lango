## ADDED Requirements

### Requirement: Interactive TUI chat on bare invocation
Running `lango` without arguments SHALL start an interactive terminal chat session using bubbletea. `lango serve` SHALL continue to work as the full gateway + channels mode.

#### Scenario: No-args launches TUI chat
- **WHEN** the user runs `lango` with no arguments on an interactive TTY
- **THEN** an interactive TUI chat session starts

#### Scenario: lango serve is unchanged
- **WHEN** the user runs `lango serve`
- **THEN** the full gateway + channels server starts as before

### Requirement: Real-time streaming responses
The TUI SHALL stream agent responses in real-time via `TurnRunner.Run()`.

#### Scenario: Streaming output displayed incrementally
- **WHEN** the agent generates a response
- **THEN** text appears incrementally in the chat viewport as tokens arrive

### Requirement: Inline tool approval prompts
Tool executions SHALL show inline approval prompts with keyboard shortcuts: `a` (allow), `s` (allow for session), `d`/`Esc` (deny).

#### Scenario: Dangerous tool triggers approval
- **WHEN** a tool with safety level Dangerous is invoked
- **THEN** an inline approval prompt appears with a/s/d key options

#### Scenario: User allows tool for session
- **WHEN** the user presses `s` on an approval prompt
- **THEN** the tool executes and future invocations of the same tool are auto-approved

### Requirement: Slash commands
The TUI SHALL support slash commands: `/help`, `/clear`, `/new`, `/model`, `/status`, `/exit`, `/quit`.

#### Scenario: /clear resets chat
- **WHEN** the user types `/clear`
- **THEN** the chat viewport is cleared and a new session starts

### Requirement: Chat history scrolling
Chat history SHALL be scrollable via PgUp/PgDn keys.

#### Scenario: Scroll up through history
- **WHEN** the user presses PgUp
- **THEN** the chat viewport scrolls up to show earlier messages

### Requirement: Markdown rendering
Completed agent responses SHALL be rendered as markdown via glamour.

#### Scenario: Code block rendered with syntax highlighting
- **WHEN** the agent response contains a fenced code block
- **THEN** it is rendered with glamour markdown formatting

### Requirement: Minimal lifecycle startup
Only Infra/Core/Buffer lifecycle components SHALL start in TUI mode (no network/automation overhead).

#### Scenario: TUI mode skips network components
- **WHEN** the app starts in local chat mode
- **THEN** `lifecycle.Registry.SetMaxPriority(PriorityBuffer)` limits startup to Infra, Core, and Buffer priorities

### Requirement: Graceful shutdown
The TUI SHALL support graceful shutdown on Ctrl+D or double Ctrl+C, and context cancellation on single Ctrl+C during streaming.

#### Scenario: Ctrl+C cancels streaming
- **WHEN** the user presses Ctrl+C while the agent is streaming
- **THEN** the current generation is cancelled but the TUI remains active

#### Scenario: Double Ctrl+C quits
- **WHEN** the user presses Ctrl+C twice in quick succession while idle
- **THEN** the TUI exits gracefully

#### Scenario: Ctrl+D exits immediately
- **WHEN** the user presses Ctrl+D
- **THEN** the TUI exits immediately

### Requirement: App mode API
The `app` package SHALL support `WithLocalChat()` option to configure local chat mode, exposing `AppMode` constants (`AppModeServer`, `AppModeLocalChat`).

#### Scenario: Local chat mode construction
- **WHEN** `app.New(boot, app.WithLocalChat())` is called
- **THEN** the app starts in local chat mode with minimal lifecycle components
