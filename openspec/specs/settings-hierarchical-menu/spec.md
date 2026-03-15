## Purpose

Define the two-level hierarchical menu navigation for the Settings TUI, replacing the flat grouped list to ensure all categories are accessible on any terminal size.

## Requirements

### Requirement: Two-level hierarchical menu navigation
The settings menu SHALL implement a two-level hierarchical navigation. Level 1 (Section List) SHALL display the 7 named sections plus Save & Cancel. Level 2 (Category Detail) SHALL display the categories within the selected section. The `MenuModel` SHALL track navigation state via `level` (levelSections or levelCategories), `activeSectionIdx`, and `sectionCursor` fields.

#### Scenario: Initial state is Level 1
- **WHEN** the menu is first created
- **THEN** the menu SHALL display Level 1 with 7 section items plus Save & Cancel (max 9 items)

#### Scenario: Enter on section transitions to Level 2
- **WHEN** user presses Enter on a section item at Level 1
- **THEN** the menu SHALL transition to Level 2 showing categories of that section
- **AND** the cursor SHALL reset to 0
- **AND** the Level 1 cursor position SHALL be saved for later restoration

#### Scenario: Enter on Save/Cancel at Level 1 sets Selected
- **WHEN** user presses Enter on Save & Exit or Cancel at Level 1
- **THEN** the menu SHALL set `Selected` to the item ID ("save" or "cancel")

#### Scenario: Enter on category at Level 2 sets Selected
- **WHEN** user presses Enter on a category item at Level 2
- **THEN** the menu SHALL set `Selected` to the category ID

#### Scenario: Esc at Level 2 returns to Level 1
- **WHEN** user presses Esc at Level 2 (not in search mode)
- **THEN** the menu SHALL return to Level 1
- **AND** the cursor SHALL be restored to the previously saved section position

### Requirement: Section items with category counts
Each section item at Level 1 SHALL display the section title and a count label showing the total number of categories in that section (e.g., "9 settings"). Section items SHALL use synthetic IDs with `__section_` prefix.

#### Scenario: Section item display
- **WHEN** the menu renders Level 1
- **THEN** each section item SHALL show its title and total category count (e.g., "Core" with "9 settings")

#### Scenario: Separator before Save/Cancel
- **WHEN** the menu renders Level 1
- **THEN** a visual separator line SHALL appear between the last named section and the Save & Exit item

### Requirement: Tab restricted to Level 2
The Tab key SHALL only toggle the Basic/Advanced filter at Level 2. At Level 1, Tab SHALL be a no-op.

#### Scenario: Tab at Level 1 is no-op
- **WHEN** user presses Tab at Level 1
- **THEN** the `showAdvanced` state SHALL remain unchanged

#### Scenario: Tab at Level 2 toggles filter
- **WHEN** user presses Tab at Level 2
- **THEN** `showAdvanced` SHALL toggle and the cursor SHALL be clamped to the visible items

### Requirement: Tab indicator at Level 2
When the menu is at Level 2, a tab indicator SHALL be rendered showing `[Basic]` and `[All]` labels. The active filter SHALL be rendered in Primary color with bold, and the inactive filter in Dim color.

#### Scenario: Tab indicator shows All active
- **WHEN** `showAdvanced` is true at Level 2
- **THEN** the tab indicator SHALL render `[All]` in Primary bold and `[Basic]` in Dim

#### Scenario: Tab indicator shows Basic active
- **WHEN** `showAdvanced` is false at Level 2
- **THEN** the tab indicator SHALL render `[Basic]` in Primary bold and `[All]` in Dim

### Requirement: Empty basic categories message
When Level 2 has no visible categories (all are advanced and `showAdvanced` is false), the menu SHALL display "No basic settings. Press Tab to show all." in muted italic text.

#### Scenario: No basic settings in section
- **WHEN** user is at Level 2 in a section with only advanced categories and `showAdvanced` is false
- **THEN** the menu SHALL display the "No basic settings" message

### Requirement: Section header at Level 2
When the menu is at Level 2, a section header SHALL be rendered above the container box showing the section title in Primary bold color with the tab indicator beside it.

#### Scenario: Section header display
- **WHEN** the menu is at Level 2 for the "Core" section
- **THEN** the header SHALL display "Core" followed by the tab indicator

#### Scenario: Section header hidden during search results
- **WHEN** the menu is at Level 2 and search results are being displayed
- **THEN** the section header SHALL NOT be shown

### Requirement: InCategoryLevel and ActiveSectionTitle public accessors
The `MenuModel` SHALL expose `InCategoryLevel() bool` and `ActiveSectionTitle() string` public methods for use by `editor.go`.

#### Scenario: InCategoryLevel at Level 1
- **WHEN** the menu is at Level 1
- **THEN** `InCategoryLevel()` SHALL return false

#### Scenario: InCategoryLevel at Level 2
- **WHEN** the menu is at Level 2
- **THEN** `InCategoryLevel()` SHALL return true

#### Scenario: ActiveSectionTitle at Level 2
- **WHEN** the menu is at Level 2 for the "Core" section
- **THEN** `ActiveSectionTitle()` SHALL return "Core"

### Requirement: Breadcrumb shows section at Level 2
When the menu is at Level 2, the editor breadcrumb SHALL display "Settings > {SectionTitle}".

#### Scenario: Breadcrumb at Level 2
- **WHEN** the menu is at Level 2 for the "Automation" section
- **THEN** the breadcrumb SHALL display "Settings > Automation"

### Requirement: Help bar adapts to navigation level
The help bar SHALL vary by navigation level. Level 1 SHALL NOT show the Tab hint. Level 2 SHALL show the Tab hint with the current filter label.

#### Scenario: Level 1 help bar
- **WHEN** the menu is at Level 1
- **THEN** the help bar SHALL show Navigate, Select, Search, and Back (no Tab)

#### Scenario: Level 2 help bar
- **WHEN** the menu is at Level 2 with `showAdvanced` true
- **THEN** the help bar SHALL show Navigate, Select, Search, Tab (Basic Only), and Back

### Requirement: Search works from both levels
The `/` search key SHALL activate global search from both Level 1 and Level 2. Search results SHALL include all categories regardless of the current section or level.

#### Scenario: Search from Level 1
- **WHEN** user presses `/` at Level 1
- **THEN** search mode SHALL activate and search all categories globally

#### Scenario: Search from Level 2
- **WHEN** user presses `/` at Level 2
- **THEN** search mode SHALL activate and search all categories globally

### Requirement: Editor Esc guard for Level 2
The editor's Esc handler at StepMenu SHALL check `InCategoryLevel()` before navigating to StepWelcome. When at Level 2, Esc SHALL be delegated to the menu's Update method for Level 2 → Level 1 transition.

#### Scenario: Esc at StepMenu Level 2 stays at menu
- **WHEN** user presses Esc at StepMenu while menu is at Level 2
- **THEN** the editor SHALL remain at StepMenu and the menu SHALL transition from Level 2 to Level 1
