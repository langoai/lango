## MODIFIED Requirements

### Requirement: Sidebar displays menu items with active highlight
The sidebar SHALL render a vertical list of menu items sourced from an `AllPageMetas()` centralized metadata table in `router.go`. The `New(items []MenuItem)` constructor SHALL accept items as a parameter instead of hardcoding them. The currently active item SHALL be visually distinguished with accent color and a left border indicator.

#### Scenario: Items sourced from AllPageMetas
- **WHEN** cockpit creates a sidebar via `sidebar.New(AllPageMetas())`
- **THEN** the sidebar SHALL display exactly 9 items in this order:
- **AND** `Mission Control`, `Chat`, `Settings`, `Tools`, `Status`, `Sessions`, `Tasks`, `Dead Letters`, and `Approvals` SHALL each have a corresponding sidebar entry
