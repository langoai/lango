## ADDED Requirements

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
