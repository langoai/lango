## MODIFIED Requirements

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
