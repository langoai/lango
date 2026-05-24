## MODIFIED Requirements

### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: Footer navigation hint appears only when another row exists
- **WHEN** the active Approvals section has two or more rows
- **THEN** the footer hint surface SHALL advertise `↑/↓` navigation

#### Scenario: Footer navigation hint hides inert keys with zero or one row
- **WHEN** the active Approvals section has zero or one row
- **THEN** the footer hint surface SHALL omit `↑/↓` navigation
