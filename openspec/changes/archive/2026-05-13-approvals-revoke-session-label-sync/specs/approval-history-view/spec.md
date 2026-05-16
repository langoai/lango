## MODIFIED Requirements

### Requirement: Approvals cockpit page
The cockpit SHALL include an Approvals page with two sections: History (top) and Grants (bottom), accessible via sidebar and ctrl+6 keybinding.

#### Scenario: Grants help uses session-scoped label for `R`
- **WHEN** the Approvals page help is rendered for the Grants section with one or more grant rows
- **THEN** the `R` binding SHALL be labeled `revoke session`
