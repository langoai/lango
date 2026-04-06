## ADDED Requirements

### Requirement: Sidebar displays menu items with active highlight
The sidebar SHALL render a vertical list of menu items sourced from an `AllPageMetas()` centralized metadata table in `router.go`. The `New(items []MenuItem)` constructor SHALL accept items as a parameter instead of hardcoding them. The currently active item SHALL be visually distinguished with accent color and a left border indicator.

#### Scenario: Render with Chat active
- **WHEN** sidebar renders with active page "Chat"
- **THEN** the Chat item SHALL display with Primary color icon, Bold label, and left accent bar
- **AND** all other items SHALL display with Muted color

#### Scenario: Items sourced from AllPageMetas
- **WHEN** cockpit creates a sidebar via `sidebar.New(AllPageMetas())`
- **THEN** the sidebar SHALL display exactly 7 items in the order: Chat, Settings, Tools, Status, Sessions, Tasks, Approvals

#### Scenario: PageID round-trip consistency
- **WHEN** any PageID's String() value is passed to PageIDFromString()
- **THEN** the original PageID SHALL be returned
- **AND** every non-Chat PageID SHALL have a corresponding entry in AllPageMetas()

### Requirement: Sidebar supports visibility toggle
The sidebar SHALL support `SetVisible(bool)` to show or hide. When hidden, `View()` SHALL return an empty string.

#### Scenario: Sidebar hidden
- **WHEN** sidebar is not visible
- **THEN** `View()` SHALL return `""`

### Requirement: Sidebar supports height adjustment
The sidebar SHALL support `SetHeight(int)` to match the terminal height. The sidebar panel SHALL fill the full terminal height.

#### Scenario: Height matches terminal
- **WHEN** `SetHeight(40)` is called
- **THEN** sidebar View SHALL render with height 40

### Requirement: Sidebar fixed width
The sidebar SHALL have a fixed width of 20 characters when fully displayed.

#### Scenario: Full width
- **WHEN** sidebar is visible
- **THEN** the rendered width SHALL be exactly 20 characters

## MODIFIED Requirements

### Requirement: Sidebar interactive navigation
The sidebar SHALL support interactive navigation when focused. When `focused=true`, up/down SHALL move the cursor and Enter SHALL emit `PageSelectedMsg`. When `focused=false`, the sidebar SHALL be display-only for keyboard events. Mouse clicks SHALL work regardless of focus state.

#### Scenario: Focused sidebar receives keys
- **WHEN** sidebar is focused and user presses down arrow
- **THEN** cursor SHALL move to the next enabled item

#### Scenario: Unfocused sidebar ignores keys
- **WHEN** sidebar is not focused and user presses down arrow
- **THEN** sidebar SHALL return unchanged (key passes through to cockpit)

#### Scenario: Enter on focused item emits PageSelectedMsg
- **WHEN** sidebar is focused and user presses Enter on an enabled item
- **THEN** sidebar SHALL return a `PageSelectedMsg{ID: item.ID}` command

#### Scenario: Disabled items skipped
- **WHEN** cursor moves via up/down
- **THEN** disabled items SHALL be skipped

#### Scenario: Enter on disabled item is no-op
- **WHEN** user presses Enter on a disabled item
- **THEN** no PageSelectedMsg SHALL be emitted

#### Scenario: Mouse click navigates regardless of focus
- **WHEN** sidebar receives tea.MouseMsg with MouseActionRelease at valid item Y coordinate
- **THEN** sidebar SHALL emit PageSelectedMsg{ID: item.ID} regardless of focused state

#### Scenario: Mouse click on disabled item is no-op
- **WHEN** sidebar receives mouse click on a disabled item
- **THEN** no PageSelectedMsg SHALL be emitted

#### Scenario: Sessions item enabled
- **WHEN** sidebar is created via New()
- **THEN** the sessions item SHALL have Disabled=false

#### Scenario: SetActive syncs cursor
- **WHEN** SetActive(id) is called
- **THEN** cursor SHALL move to the index of the matching item so that visual highlight and keyboard cursor are always aligned

#### Scenario: Page switch then Tab+Enter navigates correctly
- **WHEN** user switches page via Ctrl+N then Tab-focuses sidebar and presses Enter
- **THEN** the Enter action SHALL navigate to the same page that is visually highlighted
