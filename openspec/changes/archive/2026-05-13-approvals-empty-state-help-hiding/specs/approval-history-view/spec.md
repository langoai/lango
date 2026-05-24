## MODIFIED Requirements
### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: Help advertises both section-toggle keys readably
- **WHEN** the Approvals page help is rendered
- **THEN** it SHALL advertise both `tab` and `/` as the section-toggle keys
- **AND** the rendered help key label SHALL be `tab /`

#### Scenario: Section toggle accepts Tab and slash
- **WHEN** the operator is viewing the Approvals page
- **AND** presses `tab` or `/`
- **THEN** the page SHALL switch between the History and Grants sections

#### Scenario: Navigation help appears only when the active section has rows
- **WHEN** the active Approvals section contains one or more rows
- **THEN** the help SHALL advertise `↑/k` and `↓/j`

#### Scenario: Navigation help hides inert keys in empty or unavailable sections
- **WHEN** the active Approvals section has no rows because it is empty or unavailable
- **THEN** the help SHALL omit `↑/k` and `↓/j`
