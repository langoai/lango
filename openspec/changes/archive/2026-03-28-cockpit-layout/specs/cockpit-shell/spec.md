## ADDED Requirements

### Requirement: Cockpit root model orchestrates 2-panel layout
The cockpit `Model` SHALL compose a sidebar panel and a child panel using `lipgloss.JoinHorizontal`. When the sidebar is visible, the child panel SHALL receive `terminalWidth - sidebarWidth` as its effective width. When the sidebar is hidden, the child panel SHALL receive the full terminal width.

#### Scenario: Initial render with sidebar visible
- **WHEN** cockpit model renders with `sidebarVisible=true` and terminal width 120
- **THEN** the output SHALL be `JoinHorizontal(sidebar.View(), child.View())` with child receiving `WindowSizeMsg{Width: 100, Height: terminalHeight}`

#### Scenario: Sidebar hidden
- **WHEN** cockpit model renders with `sidebarVisible=false`
- **THEN** the output SHALL be `child.View()` only, with child receiving full terminal width

### Requirement: Consume-or-forward message delegation
The cockpit `Update()` SHALL consume only messages it handles and forward all others to the child. Consumed messages: `tea.WindowSizeMsg` (converted to reduced-width msg for child), `tea.KeyMsg` matching `Ctrl+B`. All other messages — including `ChunkMsg`, `DoneMsg`, `ErrorMsg`, `WarningMsg`, `ApprovalRequestMsg`, `SystemMsg`, remaining `KeyMsg`, `MouseMsg` — SHALL be forwarded to `child.Update(msg)`.

#### Scenario: ChunkMsg forwarded to child
- **WHEN** cockpit receives a `ChunkMsg`
- **THEN** cockpit SHALL call `child.Update(ChunkMsg)` and return the child's command

#### Scenario: Ctrl+B consumed by cockpit
- **WHEN** cockpit receives `Ctrl+B` KeyMsg
- **THEN** cockpit SHALL toggle `sidebarVisible` and NOT forward the KeyMsg to child

#### Scenario: Enter key forwarded to child
- **WHEN** cockpit receives `Enter` KeyMsg
- **THEN** cockpit SHALL forward it to `child.Update(msg)` without consuming it

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
