## MODIFIED Requirements

### Requirement: Sidebar interactive navigation
The sidebar SHALL support interactive navigation when focused. When `focused=true`, up/down SHALL move the cursor and Enter SHALL emit `PageSelectedMsg`. When `focused=false`, the sidebar SHALL be display-only for keyboard events. Mouse clicks SHALL work regardless of focus state.

#### Scenario: Focused sidebar receives keys
- **WHEN** sidebar is focused and user presses down arrow
- **THEN** cursor SHALL move to the next enabled item

#### Scenario: Unfocused sidebar ignores keys
- **WHEN** sidebar is not focused and user presses down arrow
- **THEN** sidebar SHALL return unchanged (key passes through to cockpit)

#### Scenario: Mouse click navigates regardless of focus
- **WHEN** sidebar receives tea.MouseMsg with MouseActionRelease at valid item Y coordinate
- **THEN** sidebar SHALL emit PageSelectedMsg{ID: item.ID} regardless of focused state

#### Scenario: Mouse click on disabled item is no-op
- **WHEN** sidebar receives mouse click on a disabled item
- **THEN** no PageSelectedMsg SHALL be emitted

#### Scenario: Sessions item enabled
- **WHEN** sidebar is created via New()
- **THEN** the sessions item SHALL have Disabled=false
