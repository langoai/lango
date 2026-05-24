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

#### Scenario: Navigation help appears only when another row exists
- **WHEN** the active Approvals section has two or more rows
- **THEN** the help SHALL advertise `↑/k` and `↓/j`

#### Scenario: Navigation help hides inert keys with zero or one row
- **WHEN** the active Approvals section has zero or one row
- **THEN** the help SHALL omit `↑/k` and `↓/j`

#### Scenario: Revoke help appears only when grant rows exist
- **WHEN** the Approvals page help is rendered for the Grants section with one or more grant rows
- **THEN** it SHALL advertise `r` and `R`

#### Scenario: Revoke help hides inert actions in empty grants section
- **WHEN** the Approvals page help is rendered for the Grants section with zero grant rows
- **THEN** it SHALL omit `r` and `R`
