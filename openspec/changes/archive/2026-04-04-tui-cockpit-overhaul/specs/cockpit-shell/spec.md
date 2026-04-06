## ADDED Requirements

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
