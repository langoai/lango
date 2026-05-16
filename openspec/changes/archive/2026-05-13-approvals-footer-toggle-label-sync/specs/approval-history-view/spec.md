## MODIFIED Requirements

### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: Footer advertises both section-toggle keys readably
- **WHEN** the Approvals page footer hint strip is rendered
- **THEN** it SHALL advertise both `tab` and `/` as the section-toggle keys
- **AND** the rendered footer key label SHALL be `tab /`
