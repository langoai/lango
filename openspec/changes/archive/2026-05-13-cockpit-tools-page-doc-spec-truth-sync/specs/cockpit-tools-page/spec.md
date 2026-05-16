## MODIFIED Requirements

### Requirement: Tool catalog browser with categories
ToolsPage SHALL display categories from `ToolCatalog.ListCategories()` with cursor navigation. The currently highlighted category SHALL immediately drive the tool details shown in the right panel.

#### Scenario: Browse categories
- **WHEN** ToolsPage is active
- **THEN** it SHALL display all registered categories with tool counts
- **AND** it SHALL NOT require enabled-badge rendering as part of this contract

#### Scenario: Cursor movement updates tool details
- **WHEN** the user moves the category cursor with up/down navigation
- **THEN** the right panel SHALL display tool names, descriptions, and safety levels for the currently highlighted category
- **AND** the contract SHALL NOT require an Enter key confirmation step for category selection
