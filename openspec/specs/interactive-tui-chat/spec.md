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

### Requirement: Tool and thinking transcript item kinds
The chat transcript SHALL support `itemTool` and `itemThinking` kinds in addition to existing `itemUser`, `itemAssistant`, `itemSystem`, `itemStatus`, `itemApproval`.

#### Scenario: Tool item in transcript
- **WHEN** a `ToolStartedMsg` is received
- **THEN** a new `itemTool` entry is appended to the transcript and rendered via `renderToolBlock()`

#### Scenario: Thinking item in transcript
- **WHEN** a `ThinkingStartedMsg` is received
- **THEN** a new `itemThinking` entry is appended to the transcript and rendered via `renderThinkingBlock()`

### Requirement: New Bubble Tea message types
The chat package SHALL define `ToolStartedMsg`, `ToolFinishedMsg`, `ThinkingStartedMsg`, `ThinkingFinishedMsg`, `TaskStripTickMsg`, and `PendingIndicatorTickMsg` as tea.Msg types handled in `ChatModel.Update()`.

#### Scenario: Unknown messages ignored
- **WHEN** ChatModel receives an unrecognized tea.Msg
- **THEN** it is passed through without error

### Requirement: Pending indicator during submit-to-first-event gap
The TUI SHALL show a `⏳ Working...` indicator from turn submission until the first `ChunkMsg`, `ToolStartedMsg`, or `ThinkingStartedMsg` arrives.

#### Scenario: Pending indicator appears on submit
- **WHEN** user submits a message and transitions to `stateStreaming`
- **THEN** a pending indicator tick starts and `⏳ Working...` is displayed

#### Scenario: Pending indicator dismissed
- **WHEN** the first content event (chunk, tool, or thinking) arrives
- **THEN** the pending indicator is removed from the view

### Requirement: Task strip in chat view
The `ChatModel.View()` SHALL include a task strip between the main transcript and footer when a BackgroundManager is available.

#### Scenario: Task strip rendered
- **WHEN** BackgroundManager is non-nil and has active tasks
- **THEN** the task strip appears above the footer showing task summary

#### Scenario: Task strip absent
- **WHEN** BackgroundManager is nil
- **THEN** the task strip is not rendered and takes zero height

### Requirement: Footer operational HUD
The footer SHALL display operational status: mode indicator, permission mode, model name, context budget, and pending task count. It SHALL support condensed mode for terminals narrower than 80 columns.

#### Scenario: Wide footer
- **WHEN** terminal width >= 80
- **THEN** footer shows all fields: mode, permission, model, context budget, pending tasks

#### Scenario: Narrow footer
- **WHEN** terminal width < 80
- **THEN** footer shows only essential fields (model and help keys)

### Requirement: Approval tier dispatch in stateApproving
The `ChatModel` SHALL classify incoming `ApprovalRequestMsg` by tier and dispatch rendering to `renderApprovalStrip()` (Tier 1) or `renderApprovalDialog()` (Tier 2).

#### Scenario: Tier 1 approval renders strip
- **WHEN** an approval request is classified as `TierInline`
- **THEN** `renderApprovalStrip()` is called for rendering

#### Scenario: Tier 2 approval renders dialog
- **WHEN** an approval request is classified as `TierFullscreen`
- **THEN** `renderApprovalDialog()` is called for rendering and key dispatch routes to dialog handlers

### Requirement: BackgroundManager in Deps
The chat `Deps` struct SHALL include an optional `BackgroundManager *background.Manager` field.

#### Scenario: Deps with nil BackgroundManager
- **WHEN** `Deps.BackgroundManager` is nil
- **THEN** the ChatModel initializes without task strip functionality

### Requirement: ApprovalRequestMsg carries ViewModel
The `ApprovalRequestMsg` SHALL include an `ApprovalViewModel` field alongside the existing `Request` and `Response` fields.

#### Scenario: ViewModel populated on approval request
- **WHEN** `TUIApprovalProvider.RequestApproval()` creates an `ApprovalRequestMsg`
- **THEN** the msg includes a populated `ApprovalViewModel` with tier and risk

### Requirement: ChatModel uses composite sub-models for orthogonal state
ChatModel SHALL use composite sub-model types for CPR filtering (`cprFilter`), pending indicator (`pendingIndicator`), and approval workflow (`approvalState`). These SHALL replace primitive field groups, reducing the ChatModel struct from 23 fields to 18 fields.

#### Scenario: CPR state accessed via cpr field
- **WHEN** ChatModel processes terminal response sequences
- **THEN** it SHALL delegate to `m.cpr.Filter()`, `m.cpr.Flush()`, and `m.cpr.HandleTimeout()`

#### Scenario: Pending state accessed via pending field
- **WHEN** user submits input and waits for first content event
- **THEN** ChatModel SHALL call `m.pending.Activate()` on submit and `m.pending.Dismiss()` on first content event
- **AND** `m.pending.IsActive()` and `m.pending.Elapsed()` SHALL be used for rendering

#### Scenario: Approval state accessed via approval field
- **WHEN** an approval request arrives
- **THEN** ChatModel SHALL call `m.approval.Reset(&msg)` to initialize and `m.approval.Clear()` after response
- **AND** `m.approval.pending`, `m.approval.confirmPending` SHALL be used for rendering and key handling
