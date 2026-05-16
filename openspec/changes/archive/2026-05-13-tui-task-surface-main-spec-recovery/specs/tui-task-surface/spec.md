## MODIFIED Requirements
### Requirement: Tasks page navigation
The Tasks page SHALL support keyboard navigation for selecting tasks in the table.

#### Scenario: Cursor navigation
- **WHEN** user presses `↓` on the Tasks page outside detail mode
- **THEN** the cursor moves to the next task row

#### Scenario: Detail mode help labels the current Esc action accurately
- **WHEN** the Tasks page detail panel is open
- **THEN** the `Esc` help label SHALL describe closing the detail panel rather than a generic back action

#### Scenario: Detail mode shows scroll help only when overflow exists
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content exceeds the visible detail panel height
- **THEN** the help bar SHALL advertise `↑/k` and `↓/j` as scroll actions

#### Scenario: Detail mode hides inert scroll help when no overflow exists
- **WHEN** the Tasks page detail panel is open
- **AND** the selected task detail content fits within the visible detail panel height
- **THEN** the help bar SHALL omit `↑/k` and `↓/j`

#### Scenario: List-mode detail help appears only when a task row exists
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** one or more task rows exist
- **THEN** the help bar SHALL advertise `Enter` for task detail toggling

#### Scenario: List-mode detail help hides inert Enter in empty state
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** zero task rows exist
- **THEN** the help bar SHALL omit `Enter`

#### Scenario: List-mode navigation help appears only when another task row exists
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** two or more task rows exist
- **THEN** the help bar SHALL advertise `↑/k` and `↓/j`

#### Scenario: List-mode navigation help hides inert keys in zero-or-one-row states
- **WHEN** the Tasks page help is rendered outside detail mode
- **AND** fewer than two task rows exist
- **THEN** the help bar SHALL omit `↑/k` and `↓/j`
