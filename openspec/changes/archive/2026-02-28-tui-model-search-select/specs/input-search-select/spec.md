## ADDED Requirements

### Requirement: Searchable dropdown select field type
The TUI form system MUST support an `InputSearchSelect` field type that combines text input with a filterable dropdown list.

#### Scenario: Opening the dropdown
- **WHEN** user presses Enter on a focused InputSearchSelect field
- **THEN** dropdown opens showing all options, text input clears for searching, cursor highlights current value

#### Scenario: Filtering by typing
- **WHEN** user types characters while dropdown is open
- **THEN** options are filtered by case-insensitive substring match in real-time

#### Scenario: Navigating the dropdown
- **WHEN** user presses Up/Down while dropdown is open
- **THEN** cursor moves within filtered options, clamped to list bounds

#### Scenario: Selecting an option
- **WHEN** user presses Enter while dropdown is open with a highlighted option
- **THEN** the option is selected as the field value, dropdown closes

#### Scenario: Closing without selecting
- **WHEN** user presses Esc while dropdown is open
- **THEN** dropdown closes, previous value is preserved, filter is reset

#### Scenario: Tab navigation with open dropdown
- **WHEN** user presses Tab or Shift+Tab while dropdown is open
- **THEN** dropdown closes, value is preserved, focus moves to next/previous field

#### Scenario: Dropdown display limits
- **WHEN** dropdown has more than 8 filtered options
- **THEN** only 8 are shown with scroll following cursor, remaining count shown as "... N more"
