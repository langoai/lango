## MODIFIED Requirements

### Requirement: Sidebar interactive navigation
SetActive(id) SHALL synchronize the keyboard cursor position to match the active item index, so that visual highlight and keyboard cursor are always aligned after page switches.

#### Scenario: SetActive syncs cursor
- **WHEN** SetActive("tools") is called
- **THEN** cursor SHALL move to the index of the "tools" item

#### Scenario: Page switch via Ctrl+N then Tab+Enter
- **WHEN** user switches page via Ctrl+2 then Tab-focuses sidebar and presses Enter
- **THEN** the Enter action SHALL navigate to the same page that is visually highlighted
