## MODIFIED Requirements

### Requirement: Sidebar is non-interactive in Change-1
The sidebar SHALL support interactive navigation when focused. When `focused=true`, up/down SHALL move the cursor and Enter SHALL emit `PageSelectedMsg`. When `focused=false`, the sidebar SHALL be display-only (existing Change-1 behavior).

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
