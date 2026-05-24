## Purpose

Capability spec for cockpit-tools-page. See requirements below for scope and behavior contracts.
## Requirements
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

#### Scenario: Rendered tools-page text stays plain and single-line
- **WHEN** category names, category descriptions, tool names, tool descriptions, or safety labels contain ANSI/OSC escape sequences or embedded newlines
- **THEN** the Tools page SHALL strip those control sequences
- **AND** it SHALL normalize the displayed text to a single line before rendering it
- **AND** known safety-level colors SHALL still be selected from the sanitized safety label

### Requirement: Page interface compliance
ToolsPage SHALL implement the Page interface. Activate() SHALL return nil. Deactivate() SHALL be a no-op.

#### Scenario: ToolsPage satisfies Page
- **WHEN** `var _ Page = (*ToolsPage)(nil)` is compiled
- **THEN** compilation SHALL succeed

### Requirement: Tools page degrades to an explicit empty state without a catalog
The cockpit Tools page SHALL render a stable empty state when no tool catalog is configured.

#### Scenario: Nil catalog renders explicit unavailable message
- **WHEN** ToolsPage is constructed with a nil tool catalog
- **THEN** category refresh and view rendering SHALL succeed without panic
- **AND** both panels SHALL explain that the tool catalog is not available

#### Scenario: Empty catalog renders explicit no-categories message
- **WHEN** ToolsPage is constructed with a configured catalog that has zero categories
- **THEN** both panes SHALL explain that no categories are registered

#### Scenario: Cockpit can register Tools route without catalog
- **WHEN** the cockpit starts with no tool catalog wired
- **THEN** the Tools page route SHALL still be registerable

### Requirement: Tools page help bindings match supported navigation
The cockpit Tools page SHALL expose only the key bindings it actually handles.

#### Scenario: Tools page help lists vertical navigation only when another category exists
- **WHEN** the Tools page help is rendered with two or more categories
- **THEN** it SHALL advertise `↑/k` and `↓/j`
- **AND** it SHALL NOT advertise an `Esc back` binding unless the page implements that behavior

#### Scenario: Tools page hides navigation help with fewer than two categories
- **WHEN** the Tools page help is rendered with zero or one category
- **THEN** it SHALL omit vertical navigation bindings
