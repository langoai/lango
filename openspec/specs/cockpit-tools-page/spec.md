## Purpose

Capability spec for cockpit-tools-page. See requirements below for scope and behavior contracts.

## Requirements

### Requirement: Tool catalog browser with categories
ToolsPage SHALL display categories from `ToolCatalog.ListCategories()` with cursor navigation. Selecting a category SHALL show tools from `ListTools(category)`.

#### Scenario: Browse categories
- **WHEN** ToolsPage is active
- **THEN** it SHALL display all registered categories with tool counts and enabled badges

#### Scenario: Select category shows tools
- **WHEN** user selects a category via Enter
- **THEN** the right panel SHALL display tool names, descriptions, and safety levels for that category

### Requirement: Page interface compliance
ToolsPage SHALL implement the Page interface. Activate() SHALL return nil. Deactivate() SHALL be a no-op.

#### Scenario: ToolsPage satisfies Page
- **WHEN** `var _ Page = (*ToolsPage)(nil)` is compiled
- **THEN** compilation SHALL succeed
