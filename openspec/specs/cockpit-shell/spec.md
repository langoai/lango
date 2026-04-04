## ADDED Requirements

### Requirement: Cockpit root model orchestrates 2-panel layout
The cockpit `Model` SHALL compose a sidebar panel and a child panel using `lipgloss.JoinHorizontal`. When the sidebar is visible, the child panel SHALL receive `terminalWidth - sidebarWidth` as its effective width. When the sidebar is hidden, the child panel SHALL receive the full terminal width.

#### Scenario: Initial render with sidebar visible
- **WHEN** cockpit model renders with `sidebarVisible=true` and terminal width 120
- **THEN** the output SHALL be `JoinHorizontal(sidebar.View(), child.View())` with child receiving `WindowSizeMsg{Width: 100, Height: terminalHeight}`

#### Scenario: Sidebar hidden
- **WHEN** cockpit model renders with `sidebarVisible=false`
- **THEN** the output SHALL be `child.View()` only, with child receiving full terminal width

### Requirement: Sidebar toggle triggers synthetic resize
When `Ctrl+B` toggles `sidebarVisible`, cockpit SHALL immediately send a synthetic `tea.WindowSizeMsg` to the child with the new effective width (`terminalWidth - sidebarWidth` or `terminalWidth`).

#### Scenario: Toggle sidebar on
- **WHEN** user presses `Ctrl+B` with sidebar hidden and terminal width 120
- **THEN** child SHALL receive `WindowSizeMsg{Width: 100, Height: terminalHeight}`

#### Scenario: Toggle sidebar off
- **WHEN** user presses `Ctrl+B` with sidebar visible and terminal width 120
- **THEN** child SHALL receive `WindowSizeMsg{Width: 120, Height: terminalHeight}`

### Requirement: SetProgram delegation
Cockpit SHALL expose `SetProgram(p *tea.Program)` which delegates to `child.SetProgram(p)`. Cockpit SHALL NOT expose the child model directly.

#### Scenario: Program injection
- **WHEN** caller invokes `cockpit.SetProgram(program)`
- **THEN** child's `SetProgram(program)` SHALL be called

### Requirement: childModel interface
Cockpit SHALL define a `childModel` interface: `tea.Model` + `SetProgram(*tea.Program)`. The concrete `ChatModel` SHALL satisfy this interface (compile-time verified). Test mocks SHALL implement this interface.

#### Scenario: ChatModel satisfies interface
- **WHEN** `var _ childModel = (*chat.ChatModel)(nil)` is compiled
- **THEN** compilation SHALL succeed

### Requirement: Cockpit Deps struct
Cockpit `Deps` SHALL contain: `TurnRunner *turnrunner.Runner`, `Config *config.Config`, `SessionKey string`. ApprovalProvider SHALL NOT be included in Deps.

#### Scenario: Deps fields match App struct
- **WHEN** cockpit.New(deps) is called with fields from App struct
- **THEN** all fields SHALL be directly assignable without type conversion

## MODIFIED Requirements

### Requirement: Consume-or-forward message delegation
The cockpit model's Update function SHALL consume global keys (Ctrl+1-5, Tab, Ctrl+B, Ctrl+P, Ctrl+Y) and forward all other messages to the active page or child model. Ctrl+5 is consumed to switch to the Tasks page.

#### Scenario: Ctrl+2 switches to settings
- **WHEN** user presses Ctrl+2
- **THEN** the cockpit activates the Settings page

#### Scenario: Ctrl+5 switches to tasks
- **WHEN** user presses Ctrl+5
- **THEN** the cockpit activates the Tasks page

#### Scenario: Tab toggles focus to sidebar
- **WHEN** user presses Tab
- **THEN** focus toggles between sidebar and content area

#### Scenario: PageSelectedMsg from sidebar
- **WHEN** sidebar emits `PageSelectedMsg{PageTasks}`
- **THEN** cockpit calls `switchPage(PageTasks)`

### Requirement: View dispatches to active page
View() SHALL dispatch to the active page: `child.View()` for PageChat, `pages[activePage].View()` for others. Layout composition (sidebar + main) remains unchanged.

#### Scenario: StatusPage active renders status view
- **WHEN** activePage is PageStatus
- **THEN** View() SHALL render sidebar + StatusPage.View()

### Requirement: 3-panel layout with context panel
When the context panel is visible, View() SHALL compose up to 3 panels: sidebar (left, optional), main content (center), context panel (right, optional). All components SHALL receive correct effective widths accounting for both sidebar and context panel.

#### Scenario: All three panels visible
- **WHEN** sidebarVisible=true and contextVisible=true and terminal width 120
- **THEN** child SHALL receive WindowSizeMsg{Width: 120 - 20 - 28 = 72, Height: terminalHeight}

#### Scenario: Only context panel visible
- **WHEN** sidebarVisible=false and contextVisible=true and terminal width 120
- **THEN** child SHALL receive WindowSizeMsg{Width: 120 - 28 = 92, Height: terminalHeight}

### Requirement: Mouse routing to sidebar
The cockpit SHALL route tea.MouseMsg to the sidebar when the click X coordinate is within the sidebar width. Mouse events SHALL be processed regardless of sidebar focus state.

#### Scenario: Click in sidebar region
- **WHEN** mouse click occurs at X < sidebarWidth
- **THEN** the event SHALL be forwarded to sidebar.Update(), not to child

#### Scenario: Click in content region
- **WHEN** mouse click occurs at X >= sidebarWidth
- **THEN** the event SHALL be forwarded to the active page or child

### Requirement: Clipboard copy
The cockpit SHALL support Ctrl+Y to copy the current active view content to the system clipboard.

#### Scenario: Copy chat view
- **WHEN** user presses Ctrl+Y with PageChat active
- **THEN** child.View() content SHALL be written to the system clipboard

### Requirement: Context panel toggle with synthetic resize
When Ctrl+P toggles the context panel, cockpit SHALL send synthetic WindowSizeMsg to all components with updated effective widths. Additionally, the context panel itself SHALL receive its correct width on toggle-on, even if it previously received width=0 while hidden.

#### Scenario: Toggle context panel on
- **WHEN** user presses Ctrl+P to show context panel
- **THEN** contextPanel.Start() SHALL be called and all components SHALL receive reduced width

#### Scenario: Toggle context panel off
- **WHEN** user presses Ctrl+P to hide context panel
- **THEN** contextPanel.Stop() SHALL be called and all components SHALL receive increased width

#### Scenario: First toggle after hidden initial state
- **WHEN** context panel was hidden during initial WindowSizeMsg (received width=0) and user presses Ctrl+P
- **THEN** the context panel SHALL receive WindowSizeMsg with Width=ContextPanelWidth before rendering

### Requirement: TTY Guard for TUI Commands
The root command, `cockpit` subcommand, and `chat` subcommand SHALL detect whether stdin is an interactive terminal before launching the TUI. Non-interactive environments MUST NOT attempt to start bubbletea.

#### Scenario: Root command in non-TTY environment
- **WHEN** `lango` is invoked without an interactive terminal (e.g., piped stdin, CI, `</dev/null`)
- **THEN** the command SHALL print help text and exit with code 0

#### Scenario: Cockpit subcommand in non-TTY environment
- **WHEN** `lango cockpit` is invoked without an interactive terminal
- **THEN** the command SHALL return an error: "cockpit requires an interactive terminal"

#### Scenario: Chat subcommand in non-TTY environment
- **WHEN** `lango chat` is invoked without an interactive terminal
- **THEN** the command SHALL return an error: "chat requires an interactive terminal"

#### Scenario: Normal interactive execution
- **WHEN** `lango` is invoked in an interactive terminal
- **THEN** the cockpit TUI SHALL launch normally (no behavior change)

### Requirement: Tasks page registration
The cockpit SHALL register a Tasks page at `PageTasks` (ID 5) accessible via Ctrl+5 keyboard shortcut.

#### Scenario: Ctrl+5 switches to Tasks page
- **WHEN** user presses Ctrl+5
- **THEN** the cockpit deactivates the current page and activates the Tasks page

#### Scenario: Tasks page in sidebar
- **WHEN** the sidebar is rendered
- **THEN** a "Tasks" menu entry is visible at position 5

### Requirement: BackgroundManager in cockpit Deps
The cockpit `Deps` struct SHALL include a `BackgroundManager *background.Manager` field for passing to the Tasks page and ChatModel.

#### Scenario: Deps with BackgroundManager
- **WHEN** cockpit is constructed with `Deps.BackgroundManager` set
- **THEN** the Tasks page receives the manager reference

#### Scenario: Deps without BackgroundManager
- **WHEN** cockpit is constructed with `Deps.BackgroundManager` as nil
- **THEN** the Tasks page renders a fallback message
